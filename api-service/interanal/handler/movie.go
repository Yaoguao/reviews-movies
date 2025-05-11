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

var urlDataServiceMovies = os.Getenv("DATA_SERVICE_HOST") + "/api/movies"

func (h *Handler) CreateMovieHandler(c *gin.Context) {
	var input struct {
		Title         string    `json:"title"`
		Year          int32     `json:"year"`
		Runtime       int32     `json:"runtime"`
		Genres        []string  `json:"genres"`
		CorrelationId uuid.UUID `json:"correlation_id"`
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

	err = h.producer.Producer(string(msgBytes), "movies-topic", time.Now())
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
	urlCorrelation := urlDataServiceMovies + "/by-correlation" + fmt.Sprintf("/%s", corrID)

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

func (h *Handler) GetMovieByIdHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	urlId := urlDataServiceMovies + fmt.Sprintf("/%s", idParam)

	log.Printf(urlId)
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

func (h *Handler) UpdateMovieHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	urlId := urlDataServiceMovies + fmt.Sprintf("/%s", idParam)

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

func (h *Handler) DeleteMovieHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	urlId := urlDataServiceMovies + fmt.Sprintf("/%s", idParam)

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

func (h *Handler) ListMovieHandler(c *gin.Context) {
	title := c.Query("title")
	genres := c.QueryArray("genres")
	sort := c.Query("sort")
	page := c.Query("page")
	pageSize := c.Query("page_size")

	query := url.Values{}
	if title != "" {
		query.Set("title", title)
	}
	if sort != "" {
		query.Set("sort", sort)
	}
	if page != "" {
		query.Set("page", page)
	}
	if pageSize != "" {
		query.Set("page_size", pageSize)
	}
	for _, genre := range genres {
		if genre != "" {
			query.Add("genres", genre)
		}
	}
	remoteUrl := urlDataServiceMovies + "?" + query.Encode()

	resp, err := http.Get(remoteUrl)
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

func (h *Handler) TopRatedMoviesHandler(c *gin.Context) {
	urlGet := urlDataServiceMovies + "/top"

	log.Printf(urlGet)
	resp, err := http.Get(urlGet)
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

func (h *Handler) GetWithoutReviews(c *gin.Context) {

	urlId := urlDataServiceMovies + "/without-reviews"

	log.Printf(urlId)
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

func (h *Handler) GetControversialMovies(c *gin.Context) {

	urlId := urlDataServiceMovies + "/variance"

	log.Printf(urlId)
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

func (h *Handler) GetAvgRatingByGenre(c *gin.Context) {
	urlId := urlDataServiceMovies + "/avg-rating"

	log.Printf(urlId)
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
