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

type ReviewInput struct {
	CorrelationId uuid.UUID `json:"correlation_id"`
	MovieId       uint      `json:"movie_id"`
	Rating        float32   `json:"rating"`
	Comment       string    `json:"comment"`
	Author        string    `json:"author"`
}

func (h *Handler) HandleReviewMessage(message []byte, offset kafka.Offset) error {
	var input ReviewInput

	if err := json.Unmarshal(message, &input); err != nil {
		return fmt.Errorf("json unmarshal failed: %w", err)
	}

	review := &data.Review{
		MovieId:       input.MovieId,
		Rating:        input.Rating,
		Comment:       input.Comment,
		Author:        input.Author,
		CorrelationId: input.CorrelationId,
	}

	v := validator.New()

	if data.ValidateReview(v, review); !v.Valid() {
		return fmt.Errorf("no correct valid review: %s", v.Errors)
	}

	err := h.models.Reviews.Insert(review)

	if err != nil {
		return fmt.Errorf("falied create review: %s", v.Errors)
	}

	h.logger.Info("KAFKA CREATE REVIEW OFFSET", map[string]string{
		"offset": offset.String(),
	})

	return nil
}

func (h *Handler) GetReviewByIdHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	review, err := h.models.Reviews.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": review})
}

func (h *Handler) GetReviewByCorrelation(c *gin.Context) {
	corrID, err := uuid.Parse(c.Param("correlation_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid correlation_id"})
		return
	}
	review, err := h.models.Reviews.GetReviewByCorrelation(corrID)
	if err != nil {
		c.JSON(http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"review": review})
}

func (h *Handler) UpdateReviewHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.models.Reviews.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	var input struct {
		Rating  *float32 `json:"rating"`
		Comment *string  `json:"comment"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	updates := &data.Review{
		Model:         existing.Model,
		MovieId:       existing.MovieId,
		Rating:        existing.Rating,
		Author:        existing.Author,
		CorrelationId: existing.CorrelationId,
		Version:       existing.Version,
		Comment:       existing.Comment,
	}
	if input.Rating != nil {
		updates.Rating = *input.Rating
	}
	if input.Comment != nil {
		updates.Comment = *input.Comment
	}

	v := validator.New()
	data.ValidateReview(v, updates)
	if !v.Valid() {
		c.JSON(http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = h.models.Reviews.Update(updates)
	if err != nil {
		if errors.Is(err, models.ErrEditConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "edit conflict"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": updates})
}

func (h *Handler) DeleteReviewHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.models.Reviews.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operation delete successfully completed"})
}

func (h *Handler) ListReviewHandler(c *gin.Context) {
	movieIDStr := c.DefaultQuery("movie_id", "0")
	movieID, err := strconv.ParseUint(movieIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie_id"})
		return
	}

	author := c.DefaultQuery("author", "")
	ratingStr := c.DefaultQuery("rating", "0")
	rating, err := strconv.ParseFloat(ratingStr, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rating"})
		return
	}

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

	sort := c.DefaultQuery("sort", "id")

	filters := data.Filters{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		SortSafelist: []string{
			"id", "rating", "author", "-id", "-rating", "-author",
		},
	}

	reviews, metadata, err := h.models.Reviews.GetAll(
		c.Request.Context(),
		uint(movieID),
		author,
		float32(rating),
		filters,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews":  reviews,
		"metadata": metadata,
	})
}
