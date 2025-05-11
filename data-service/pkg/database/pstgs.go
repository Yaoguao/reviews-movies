package database

import (
	"data-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenDB(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DB.Dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
