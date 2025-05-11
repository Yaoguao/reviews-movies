package handler

import (
	"data-service/internal/data"
	"data-service/internal/models"
	"data-service/internal/validator"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
)

type MovieInput struct {
	Title         string    `json:"title" binding:"required"`
	Year          int32     `json:"year" binding:"required"`
	Runtime       int32     `json:"runtime"`
	Genres        []string  `json:"genres"`
	CorrelationId uuid.UUID `json:"correlation_id"`
}

type MovieDTO struct {
	ID      uint     `json:"id"`
	Title   string   `json:"title"`
	Year    int32    `json:"year"`
	Runtime int32    `json:"runtime"`
	Genres  []string `json:"genres"`
	Version int32    `json:"version"`
}

type MovieUpdateInput struct {
	Title   *string  `json:"title,omitempty"`
	Year    *int32   `json:"year,omitempty"`
	Runtime *int32   `json:"runtime,omitempty"`
	Genres  []string `json:"genres,omitempty"`
}

func (h *Handler) HandleMovieMessage(message []byte, offset kafka.Offset) error {
	var input MovieInput

	if err := json.Unmarshal(message, &input); err != nil {
		return fmt.Errorf("json unmarshal failed: %w", err)
	}

	movie := &data.Movie{
		Title:         input.Title,
		Year:          input.Year,
		Runtime:       input.Runtime,
		Genres:        input.Genres,
		CorrelationId: input.CorrelationId,
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		return fmt.Errorf("no correct valid movie: %s", v.Errors)
	}

	err := h.models.Movies.Insert(movie)

	if err != nil {
		return fmt.Errorf("falied create movie: %s", v.Errors)
	}

	h.logger.PrintInfo("KAFKA CREATE MOVIE OFFSET", map[string]string{
		"offset": offset.String(),
	})

	return nil
}

func (h *Handler) GetMovieByIdHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	movie, err := h.models.Movies.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"movie": movie})
}

func (h *Handler) GetMovieByCorrelation(c *gin.Context) {
	corrID, err := uuid.Parse(c.Param("correlation_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid correlation_id"})
		return
	}
	movie, err := h.models.Movies.GetMovieByCorrelation(corrID)
	if err != nil {
		c.JSON(http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"movie": movie})
}

func (h *Handler) UpdateMovieHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.models.Movies.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	var input MovieUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	updates := &data.Movie{
		Model:   existing.Model,
		Title:   existing.Title,
		Year:    existing.Year,
		Runtime: existing.Runtime,
		Genres:  existing.Genres,
		Version: existing.Version,
	}
	if input.Title != nil {
		updates.Title = *input.Title
	}
	if input.Year != nil {
		updates.Year = *input.Year
	}
	if input.Runtime != nil {
		updates.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		updates.Genres = input.Genres
	}

	v := validator.New()
	data.ValidateMovie(v, updates)
	if !v.Valid() {
		c.JSON(http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = h.models.Movies.Update(updates)
	if err != nil {
		if errors.Is(err, models.ErrEditConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "edit conflict"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update movie"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"movie": updates})
}

func (h *Handler) DeleteMovieHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.models.Movies.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operation delete successfully completed"})
}

func (h *Handler) ListMovieHandler(c *gin.Context) {
	//var input struct {
	//	Title  string
	//	Genres []string
	//	data.Filters
	//}
	//
	//v := validator.New()

	title := c.DefaultQuery("title", "")
	genres := c.QueryArray("genres")
	sort := c.DefaultQuery("sort", "id")

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}
	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size"})
		return
	}

	filters := data.Filters{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		SortSafelist: []string{
			"id", "title", "year", "runtime",
			"-id", "-title", "-year", "-runtime",
		},
	}

	movies, metadata, err := h.models.Movies.GetAll(c.Request.Context(), title, genres, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"movies":   movies,
		"metadata": metadata,
	})
}

func (h *Handler) TopRatedMoviesHandler(c *gin.Context) {
	res, err := h.models.Movies.GetTopRated(c.Request.Context(), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch top rated movies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"top_movies": res})
}

func (h *Handler) GetWithoutReviews(c *gin.Context) {
	res, err := h.models.Movies.GetWithoutReviews(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"movies_without_reviews": res})
}

func (h *Handler) GetControversialMovies(c *gin.Context) {
	res, err := h.models.Movies.GetControversialMovies(c.Request.Context(), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"movies_variance": res})
}

func (h *Handler) GetAvgRatingByGenre(c *gin.Context) {
	res, err := h.models.Movies.GetAvgRatingByGenre(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rating_average": res})
}
