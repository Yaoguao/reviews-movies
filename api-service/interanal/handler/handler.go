package handler

import (
	"github.com/gin-gonic/gin"
	"reviews-movies/api-service/interanal/kafka"
)

type Handler struct {
	producer *kafka.Producer
}

func NewHandler(producer *kafka.Producer) *Handler {
	return &Handler{producer: producer}
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
			movies.GET("/top/", h.TopRatedMoviesHandler)
			movies.GET("/without-reviews", h.GetWithoutReviews)
			movies.GET("/variance", h.GetControversialMovies)
			movies.GET("/avg-rating", h.GetAvgRatingByGenre)
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
