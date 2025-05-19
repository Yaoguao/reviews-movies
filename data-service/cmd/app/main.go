package main

import (
	"context"
	"data-service/internal/config"
	"data-service/internal/data"
	"data-service/internal/handler"
	"data-service/internal/kafka"
	logger2 "data-service/internal/logger"
	"data-service/internal/models"
	"data-service/pkg/database"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

	logger.Debug(fmt.Sprintf("Config: %+v", *cfg))
	db, err := database.OpenDB(cfg)
	if err != nil {
		logger.Info(err.Error(), nil)
	}
	if err := db.AutoMigrate(&data.Movie{}, &data.Review{}); err != nil {
		logger.Info(err.Error(), nil)
	}
	defer func() {
		conn, _ := db.DB()
		conn.Close()
	}()
	logger.Info("database connection pool established", nil)

	app := &Application{config: cfg, logger: logger}
	appModels := models.NewModels(db)
	ginHandler := handler.NewHandler(appModels, logger)

	movieConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Address,
		cfg.Kafka.Topics.Movie,
		cfg.Kafka.ConsumerGroup.Movie,
	)
	if err != nil {
		logger.Info(err.Error(), nil)
		os.Exit(1)
	}

	reviewConsumer, err := kafka.NewConsumer(
		cfg.Kafka.Address,
		cfg.Kafka.Topics.Review,
		cfg.Kafka.ConsumerGroup.Review,
	)
	if err != nil {
		logger.Info(err.Error(), nil)
		os.Exit(1)
	}

	go movieConsumer.StartWithFunc(ginHandler.HandleMovieMessage)
	go reviewConsumer.StartWithFunc(ginHandler.HandleReviewMessage)

	if err := app.serve(ginHandler.Routes()); err != nil {
		logger.Info(err.Error(), nil)
	}
	if err := movieConsumer.Stop(); err != nil {
		app.logger.Info(err.Error(), nil)
	}
	if err := reviewConsumer.Stop(); err != nil {
		app.logger.Info(err.Error(), nil)
	}
}

func (app *Application) serve(router http.Handler) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.Port),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// канал для ошибок shutdown
	shutdownErr := make(chan error, 1)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		app.logger.Info("caught signal, shutting down", map[string]string{"signal": s.String()})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(ctx)
	}()

	app.logger.Info("starting Gin server", map[string]string{
		"env":  app.config.Env,
		"addr": srv.Addr,
	})

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownErr; err != nil {
		return err
	}

	app.logger.Info("server stopped", map[string]string{"addr": srv.Addr})
	return nil
}

//logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
//if err != nil {
//log.Fatalf("could not open log file: %v", err)
//}
