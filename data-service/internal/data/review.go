package data

import (
	"data-service/internal/validator"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
	"strings"
)

type Review struct {
	gorm.Model
	CorrelationId uuid.UUID              `gorm:"type:uuid;not null;uniqueIndex" json:"correlation_id"`
	MovieId       uint                   `gorm:"not null" json:"movie_id"`
	Rating        float32                `gorm:"type:decimal(2,1);check:rating >= 0.5 AND rating <= 5.0" json:"rating"`
	Comment       string                 `gorm:"type:text;not null" json:"comment"`
	Author        string                 `gorm:"size:50;not null" json:"author"`
	Version       optimisticlock.Version `gorm:"default:1" json:"version"`

	Movie Movie `gorm:"foreignKey:MovieId;references:ID" json:"-"`
}

func ValidateReview(v *validator.Validator, review *Review) {
	v.Check(review.CorrelationId != uuid.Nil, "correlation_id", "must be a valid UUID")

	v.Check(review.MovieId != 0, "movie_id", "must be provided")

	v.Check(review.Rating >= 0.5 && review.Rating <= 5.0, "rating", "must be between 0.5 and 5.0")

	v.Check(strings.TrimSpace(review.Comment) != "", "comment", "must be provided")
	v.Check(len(review.Comment) <= 1000, "comment", "must not be more than 1000 characters")

	v.Check(strings.TrimSpace(review.Author) != "", "author", "must be provided")
	v.Check(len(review.Author) <= 50, "author", "must not be more than 50 characters")
}
