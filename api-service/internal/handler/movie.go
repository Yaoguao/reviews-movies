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

func (h *Handler) CreateMovieHandler(c *gin.Context) {
	var input struct {
		Title         string    `json:"title"`
		Year          int32     `json:"year"`
		Runtime       int32     `json:"runtime"`
		Genres        []string  `json:"genres"`
		CorrelationId uuid.UUID `json:"correlation_id" binding:"-"`
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
	h.logger.Debug(input.Title, nil)

	err = h.producer.Produce(string(msgBytes), h.cfg.Kafka.Topics.Movie, input.CorrelationId.String(), time.Now())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "correlation_id": input.CorrelationId})
}

func (h *Handler) GetMovieByCorrelation(c *gin.Context) {
	corrID, err := uuid.Parse(c.Param("correlation_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid correlation_id"})
		return
	}
	body, status, contentType, err := h.moviesClient.GetByCorrelationID(c, corrID.String())
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) GetMovieByIdHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	body, status, contentType, err := h.moviesClient.GetByID(c, id)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) UpdateMovieHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	payload, _ := io.ReadAll(c.Request.Body)

	body, status, contentType, err := h.moviesClient.UpdateMovie(c, id, payload, nil)

	c.Data(status, contentType, body)
}

func (h *Handler) DeleteMovieHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	body, status, contentType, err := h.moviesClient.DeleteMovie(c, id)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) ListMovieHandler(c *gin.Context) {
	title := c.Query("title")
	genres := c.QueryArray("genres")
	sort := c.Query("sort")
	page := c.Query("page")
	pageSize := c.Query("page_size")

	body, status, contentType, err := h.moviesClient.GetListMovie(c, title, sort, page, pageSize, genres)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) GetTopRatedMoviesHandler(c *gin.Context) {
	body, status, contentType, err := h.moviesClient.GetTopRatedMovies(c)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) GetWithoutReviewsHandler(c *gin.Context) {
	body, status, contentType, err := h.moviesClient.GetWithoutReviews(c)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) GetControversialMoviesHandler(c *gin.Context) {
	body, status, contentType, err := h.moviesClient.GetControversialMovies(c)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}

func (h *Handler) GetAvgRatingByGenreHandler(c *gin.Context) {
	body, status, contentType, err := h.moviesClient.GetAvgRatingByGenre(c)
	if err != nil {
		h.logger.Error("ERROR: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "¯\\_(ツ)_/¯"})
		return
	}
	c.Data(status, contentType, body)
}
