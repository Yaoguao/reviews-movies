package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port int
	Env  string
	DB   struct {
		Dsn string
	}
	Kafka struct {
		Address []string
		Topics  struct {
			Movie  string
			Review string
		}
		ConsumerGroup struct {
			Movie  string
			Review string
		}
	}
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(getEnv("DATA_SERVICE_PORT", "8081"))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port: port,
		Env:  getEnv("DATA_SERVICE_ENV", "development"),
	}

	cfg.DB.Dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_SSLMODE"),
	)

	cfg.Kafka.Address = []string{
		os.Getenv("BROKER1_HOST"),
		os.Getenv("BROKER2_HOST"),
		os.Getenv("BROKER3_HOST"),
	}

	cfg.Kafka.Topics.Movie = getEnv("MOVIES_TOPIC", "movies-topic")
	cfg.Kafka.Topics.Review = getEnv("REVIEWS_TOPIC", "reviews-topic")

	cfg.Kafka.ConsumerGroup.Movie = getEnv("MOVIES_CONSUMER_GROUP", "movies-group")
	cfg.Kafka.ConsumerGroup.Review = getEnv("REVIEWS_CONSUMER_GROUP", "reviews-group")

	return cfg, nil

}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
