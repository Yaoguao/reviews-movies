package models

import (
	"context"
	"data-service/internal/data"
	dto2 "data-service/internal/data/dto"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
)

type MovieModel struct {
	DB *gorm.DB
}

func (m *MovieModel) Insert(movie *data.Movie) error {
	res := m.DB.Create(movie)

	if res.Error != nil {
		switch {
		case strings.Contains(res.Error.Error(), "duplicate key value violates unique constraint"):
			return errors.New("duplicate key value violates unique constraint")
		default:
			return res.Error
		}
	}
	return nil
}

func (m *MovieModel) GetMovieByCorrelation(corrId uuid.UUID) (*data.Movie, error) {
	var movie data.Movie

	err := m.DB.
		Where("correlation_id = ?", corrId).
		First(&movie).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &movie, nil
}

func (m *MovieModel) Get(id uint) (*data.Movie, error) {
	var movie data.Movie
	err := m.DB.First(&movie, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &movie, nil
}

func (m *MovieModel) Update(movie *data.Movie) error {
	result := m.DB.
		Model(&data.Movie{}).
		Where("id = ? AND version = ?", movie.ID, movie.Version).
		Updates(movie)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrEditConflict
	}
	return nil
}

func (m *MovieModel) Delete(id uint) error {
	result := m.DB.Delete(&data.Movie{}, id)
	return result.Error
}

func (m *MovieModel) GetAll(ctx context.Context, title string, genres []string, filters data.Filters) ([]*data.Movie, data.Metadata, error) {
	var (
		movies       []*data.Movie
		totalRecords int64
	)

	db := m.DB.WithContext(ctx).Model(&data.Movie{})

	if title != "" {
		tsQuery := strings.TrimSpace(title)
		db = db.Where(
			"to_tsvector('simple', title) @@ plainto_tsquery('simple', ?)",
			tsQuery,
		)
	}

	if len(genres) > 0 {
		jsonGenres, err := json.Marshal(genres)
		if err != nil {
			return nil, data.Metadata{}, err
		}
		db = db.Where("genres @> ?", jsonGenres)
	}

	if err := db.Count(&totalRecords).Error; err != nil {
		return nil, data.Metadata{}, err
	}

	sortExpr := filters.SortColumn() + " " + filters.SortDirection() + ", id ASC"
	db = db.Order(sortExpr)

	db = db.Limit(filters.Limit()).Offset(filters.Offset())

	if err := db.Find(&movies).Error; err != nil {
		return nil, data.Metadata{}, err
	}

	metadata := data.CalculateMetadata(int(totalRecords), filters.Page, filters.PageSize)

	return movies, metadata, nil
}

func (m *MovieModel) GetTopRated(ctx context.Context, limit int) ([]dto2.MovieRating, error) {
	var results []dto2.MovieRating

	err := m.DB.WithContext(ctx).
		Table("movies AS m").
		Select("m.id, m.title, m.year, m.runtime, AVG(r.rating) AS avg_rating").
		Joins("JOIN reviews r ON m.id = r.movie_id").
		Group("m.id").
		Order("avg_rating DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (m *MovieModel) GetWithoutReviews(ctx context.Context) ([]dto2.MovieRating, error) {
	var results []dto2.MovieRating

	err := m.DB.WithContext(ctx).
		Table("movies AS m").
		Select("m.id, m.title, m.year, m.runtime").
		Joins("LEFT JOIN reviews r ON m.id = r.movie_id").
		Where("r.id IS NULL").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (m *MovieModel) GetControversialMovies(ctx context.Context, limit int) ([]dto2.MovieVariance, error) {
	var result []dto2.MovieVariance

	err := m.DB.WithContext(ctx).Raw(`
		SELECT
			m.id,
			m.title,
			m.year,
			m.runtime,
			VAR_SAMP(r.rating) AS variance
		FROM movies m
		JOIN reviews r ON m.id = r.movie_id
		GROUP BY m.id
		ORDER BY variance DESC
		LIMIT ?
	`, limit).Scan(&result).Error

	return result, err
}

func (m *MovieModel) GetAvgRatingByGenre(ctx context.Context) ([]dto2.GenreRating, error) {
	var result []dto2.GenreRating

	err := m.DB.WithContext(ctx).Raw(`
		SELECT
			jsonb_array_elements_text(genres) AS genre,
			AVG(r.rating) AS avg_rating
		FROM movies m
		JOIN reviews r ON m.id = r.movie_id
		GROUP BY genre
		ORDER BY avg_rating DESC
	`).Scan(&result).Error

	return result, err
}
