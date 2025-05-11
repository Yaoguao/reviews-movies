package data

import (
	"data-service/internal/validator"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
	"time"
)

type Movie struct {
	gorm.Model
	CorrelationId uuid.UUID              `gorm:"type:uuid;not null;uniqueIndex" json:"correlation_id"`
	Title         string                 `gorm:"size:100;not null" json:"title"`
	Year          int32                  `json:"year,omitempty"`
	Runtime       int32                  `json:"runtime,omitempty"`
	Genres        Genres                 `gorm:"type:jsonb" json:"genres"`
	Version       optimisticlock.Version `gorm:"default:1" json:"version"`
}

type Genres []string

func (g Genres) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *Genres) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}
	return json.Unmarshal(bytes, g)
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) != 0, "genres", "must be provided")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
