package handler

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"reviews-movies/api-service/config"
	"reviews-movies/api-service/interanal/apiclient"
	"reviews-movies/api-service/interanal/apiclient/movies"
	"reviews-movies/api-service/interanal/apiclient/reviews"
	"reviews-movies/api-service/interanal/kafka"
)

type Handler struct {
	logger   *slog.Logger
	cfg      *config.Config
	producer *kafka.Producer

	moviesClient  *movies.Client
	reviewsClient *reviews.Client
}

func NewHandler(producer *kafka.Producer, logger *slog.Logger, cfg *config.Config) *Handler {
	restyCli := apiclient.NewBaseClient(cfg.Services.DataServiceHost, 5)

	moviesCli := movies.NewMovieClient(restyCli)
	reviewsCli := reviews.NewReviewClient(restyCli)

	return &Handler{
		producer:      producer,
		logger:        logger,
		cfg:           cfg,
		moviesClient:  moviesCli,
		reviewsClient: reviewsCli,
	}
}

func (h *Handler) Routes() *gin.Engine {
	router := gin.New()
	api := router.Group("/api")
	{
		movies := api.Group("/movies")
		{
			movies.GET("/:id", h.GetMovieByIdHandler)
			movies.POST("/", h.CreateMovieHandler)
			movies.PATCH("/:id", h.UpdateMovieHandler)
			movies.DELETE("/:id", h.DeleteMovieHandler)
			movies.GET("/", h.ListMovieHandler)
			movies.GET("/by-correlation/:correlation_id", h.GetMovieByCorrelation)
			movies.GET("/top/", h.GetTopRatedMoviesHandler)
			movies.GET("/without-reviews", h.GetWithoutReviewsHandler)
			movies.GET("/variance", h.GetControversialMoviesHandler)
			movies.GET("/avg-rating", h.GetAvgRatingByGenreHandler)
		}
		reviews := api.Group("/reviews")
		{
			reviews.POST("/", h.CreateReviewHandler)
			reviews.GET("/:id", h.GetReviewByIdHandler)
			reviews.PATCH("/:id", h.UpdateReviewHandler)
			reviews.DELETE("/:id", h.DeleteReviewHandler)
			reviews.GET("/", h.ListReviewHandler)
			reviews.GET("/by-correlation/:correlation_id", h.GetReviewByCorrelation)
		}
	}

	return router
}
