package handler

import (
	"data-service/internal/jsonlog"
	"data-service/internal/models"
	"fmt"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	models *models.Models
	logger *jsonlog.Logger
}

func NewHandler(models *models.Models, logger *jsonlog.Logger) *Handler {
	return &Handler{
		models: models,
		logger: logger,
	}
}

func (h *Handler) Routes() *gin.Engine {
	router := gin.New()

	router.Use(gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		h.logger.PrintInfo("request", map[string]string{
			"method": p.Method, "path": p.Path,
			"status": fmt.Sprint(p.StatusCode), "latency": p.Latency.String(),
		})
		return ""
	}))
	router.Use(gin.Recovery())

	api := router.Group("/api")
	{
		movies := api.Group("/movies")
		{
			movies.GET("/:id", h.GetMovieByIdHandler)
			movies.PATCH("/:id", h.UpdateMovieHandler)
			movies.DELETE("/:id", h.DeleteMovieHandler)
			movies.GET("/", h.ListMovieHandler)
			movies.GET("/by-correlation/:correlation_id", h.GetMovieByCorrelation)
			movies.GET("/top/", h.TopRatedMoviesHandler)
			movies.GET("/without-reviews", h.GetWithoutReviews)
			movies.GET("/variance", h.GetControversialMovies)
			movies.GET("/avg-rating", h.GetAvgRatingByGenre)
		}
		reviews := api.Group("/reviews")
		{
			reviews.GET("/:id", h.GetReviewByIdHandler)
			reviews.PATCH("/:id", h.UpdateReviewHandler)
			reviews.DELETE("/:id", h.DeleteReviewHandler)
			reviews.GET("/", h.ListReviewHandler)
			reviews.GET("/by-correlation/:correlation_id", h.GetReviewByCorrelation)
		}
	}

	return router
}
