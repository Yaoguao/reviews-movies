package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var urlDataServiceReviews = os.Getenv("DATA_SERVICE_HOST") + "/api/reviews"

func (h *Handler) CreateReviewHandler(c *gin.Context) {
	var input struct {
		CorrelationId uuid.UUID `json:"correlation_id"`
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
	fmt.Println(input)
	fmt.Println(msgBytes)

	err = h.producer.Producer(string(msgBytes), "reviews-topic", time.Now())
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

	urlId := urlDataServiceReviews + fmt.Sprintf("/%s", idParam)

	resp, err := http.Get(urlId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to contact data service"})
		log.Printf(err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func (h *Handler) GetReviewByCorrelation(c *gin.Context) {
	corrID, err := uuid.Parse(c.Param("correlation_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid correlation_id"})
		return
	}
	urlCorrelation := urlDataServiceReviews + "/by-correlation" + fmt.Sprintf("/%s", corrID)

	log.Printf(urlCorrelation)
	resp, err := http.Get(urlCorrelation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to contact data service"})
		log.Printf(err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}

func (h *Handler) UpdateReviewHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	urlId := urlDataServiceReviews + fmt.Sprintf("/%s", idParam)

	req, err := http.NewRequest(http.MethodPatch, urlId, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	req.Header = c.Request.Header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to contact data service"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func (h *Handler) DeleteReviewHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	urlId := urlDataServiceReviews + fmt.Sprintf("/%s", idParam)

	req, err := http.NewRequest(http.MethodDelete, urlId, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	req.Header = c.Request.Header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to contact data service"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func (h *Handler) ListReviewHandler(c *gin.Context) {
	params := url.Values{}
	if movieID := c.Query("movie_id"); movieID != "" {
		params.Set("movie_id", movieID)
	}
	if author := c.Query("author"); author != "" {
		params.Set("author", author)
	}
	if rating := c.Query("rating"); rating != "" {
		params.Set("rating", rating)
	}
	if page := c.Query("page"); page != "" {
		params.Set("page", page)
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		params.Set("page_size", pageSize)
	}
	if sort := c.Query("sort"); sort != "" {
		params.Set("sort", sort)
	}

	fullURL := fmt.Sprintf("%s?%s", urlDataServiceReviews, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to contact data service"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}
