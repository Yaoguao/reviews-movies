package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reviews-movies/api-service/config"
	"reviews-movies/api-service/interanal/handler"
	"reviews-movies/api-service/interanal/kafka"
	logger2 "reviews-movies/api-service/interanal/logger"
	"strings"
	"syscall"
	"time"
)

type Application struct {
	config *config.Config
	logger *slog.Logger
}

func main() {
	cfg, _ := config.LoadConfig()

	logger := logger2.InitLogger(cfg.Env, os.Stdout)

	app := &Application{
		config: cfg,
		logger: logger,
	}

	logger.Debug("Kafka Address: "+strings.Join(cfg.Kafka.Address, ", "), nil)

	producer, err := kafka.NewProducer(cfg.Kafka.Address)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	newHandler := handler.NewHandler(producer, logger, cfg)
	app.serve(newHandler.Routes())
	log.Fatal(err, nil)
}

func (app *Application) serve(handler http.Handler) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.Port),
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.logger.Info("caught signal", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.Info("starting server", map[string]string{
		"env":  app.config.Env,
		"addr": srv.Addr,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
