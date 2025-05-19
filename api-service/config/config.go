package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port    int
	Env     string
	Limiter struct {
		Rps     float64
		Burst   int
		Enabled bool
	}
	Kafka struct {
		Address []string
		Topics  struct {
			Movie  string
			Review string
		}
	}
	Services struct {
		DataServiceHost string
	}
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(getEnv("API_SERVICE_PORT", "8082"))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port: port,
		Env:  getEnv("API_SERVICE_ENV", "development"),
	}

	cfg.Services.DataServiceHost = os.Getenv("DATA_SERVICE_HOST")

	cfg.Kafka.Address = []string{
		os.Getenv("BROKER1_HOST"),
		os.Getenv("BROKER2_HOST"),
		os.Getenv("BROKER3_HOST"),
	}

	cfg.Kafka.Topics.Movie = getEnv("MOVIES_TOPIC", "movies-topic")
	cfg.Kafka.Topics.Review = getEnv("REVIEWS_TOPIC", "reviews-topic")

	return cfg, nil

}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
