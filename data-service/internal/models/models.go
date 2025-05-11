package models

import (
	"errors"
	"gorm.io/gorm"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateKey   = errors.New("duplicate key value violates unique constraint")
)

type Models struct {
	Movies  *MovieModel
	Reviews *ReviewModel
}

func NewModels(db *gorm.DB) *Models {
	return &Models{
		Movies:  &MovieModel{DB: db},
		Reviews: &ReviewModel{DB: db},
	}
}
