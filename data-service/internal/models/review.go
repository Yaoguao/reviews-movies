package models

import (
	"context"
	"data-service/internal/data"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
)

type ReviewModel struct {
	DB *gorm.DB
}

func (r *ReviewModel) Insert(review *data.Review) error {
	var movie data.Movie
	err := r.DB.First(&movie, review.MovieId).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return fmt.Errorf("movie with id %d not found", review.MovieId)
		default:
			return err
		}
	}

	if err := r.DB.Create(review).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ErrDuplicateKey
		}
		return err
	}

	return nil
}

func (m *ReviewModel) GetReviewByCorrelation(corrId uuid.UUID) (*data.Review, error) {
	var review data.Review

	err := m.DB.
		Where("correlation_id = ?", corrId).
		First(&review).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &review, nil
}

func (m *ReviewModel) Get(id uint) (*data.Review, error) {
	var review data.Review
	err := m.DB.First(&review, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &review, nil
}

func (m *ReviewModel) Update(review *data.Review) error {
	result := m.DB.
		Model(&data.Review{}).
		Where("id = ? AND version = ?", review.ID, review.Version).
		Updates(review)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrEditConflict
	}
	return nil
}

func (m *ReviewModel) Delete(id uint) error {
	result := m.DB.Delete(&data.Review{}, id)
	return result.Error
}

func (m *ReviewModel) GetAll(ctx context.Context, movieId uint, author string, rating float32, filters data.Filters) ([]*data.Review, data.Metadata, error) {
	var (
		reviews      []*data.Review
		totalRecords int64
	)

	db := m.DB.WithContext(ctx).Model(&data.Review{})

	if author != "" {
		tsQuery := strings.TrimSpace(author)
		db = db.Where(
			"to_tsvector('simple', title) @@ plainto_tsquery('simple', ?)",
			tsQuery,
		)
	}

	if movieId > 0 {
		db = db.Where("movie_id = ?", movieId)
	}

	if rating > 0 {
		db = db.Where("rating = ?", rating)
	}

	if err := db.Count(&totalRecords).Error; err != nil {
		return nil, data.Metadata{}, err
	}

	sortExpr := filters.SortColumn() + " " + filters.SortDirection() + ", id ASC"
	db = db.Order(sortExpr)

	db = db.Limit(filters.Limit()).Offset(filters.Offset())

	if err := db.Find(&reviews).Error; err != nil {
		return nil, data.Metadata{}, err
	}

	metadata := data.CalculateMetadata(int(totalRecords), filters.Page, filters.PageSize)

	return reviews, metadata, nil
}
