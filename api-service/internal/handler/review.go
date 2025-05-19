package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) CreateReviewHandler(c *gin.Context) {
	var input struct {
		CorrelationId uuid.UUID `json:"correlation_id" binding:"-"`
		MovieId       uint      `json:"movie_id"`
		Rating        float32   `json:"rating"`
		Comment       string    `json:"comment"`
		Author        string    `json:"author"`
	}

	input.CorrelationId = uuid.New()

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	msgBytes, err := json.Marshal(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize message"})
		return
	}
	h.logger.Debug(input.CorrelationId.String(), nil)

	key := strconv.Itoa(int(input.MovieId))

	err = h.producer.Produce(string(msgBytes), h.cfg.Kafka.Topics.Review, key, time.Now())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "correlation_id": input.CorrelationId})
}

func (h *Handler) GetReviewByIdHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	body, status, contentType, err := h.reviewsClient.GetByID(c, id)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) GetReviewByCorrelation(c *gin.Context) {
	corrID, err := uuid.Parse(c.Param("correlation_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid correlation_id"})
		return
	}

	body, status, contentType, err := h.reviewsClient.GetByCorrelationID(c, corrID.String())
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) UpdateReviewHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	payload, _ := io.ReadAll(c.Request.Body)

	body, status, contentType, err := h.reviewsClient.UpdateReview(c, id, payload, nil)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) DeleteReviewHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	body, status, contentType, err := h.reviewsClient.DeleteReview(c, id)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) ListReviewHandler(c *gin.Context) {
	movieID := c.Query("movie_id")
	author := c.Query("author")
	rating := c.Query("rating")
	page := c.Query("page")
	pageSize := c.Query("page_size")
	sort := c.Query("sort")

	body, status, contentType, err := h.reviewsClient.GetListReview(c, movieID, author, rating, sort, page, pageSize)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}
