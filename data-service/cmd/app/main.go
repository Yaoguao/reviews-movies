package main

import (
	"context"
	"data-service/internal/config"
	"data-service/internal/data"
	"data-service/internal/handler"
	"data-service/internal/jsonlog"
	"data-service/internal/kafka"
	"data-service/internal/models"
	"data-service/pkg/database"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	config config.Config
	logger *jsonlog.Logger
}

var address = []string{
	os.Getenv("BROKER1_HOST"),
	os.Getenv("BROKER2_HOST"),
	os.Getenv("BROKER3_HOST"),
}

var (
	moviesTopic  = "movies-topic"
	reviewsTopic = "reviews-topic"
	moviesGroup  = "movies-group"
	reviewsGroup = "reviews-group"
)

func main() {
	var cfg config.Config
	flag.IntVar(&cfg.Port, "port", 8081, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	cfg.DB.Dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_SSLMODE"),
	)
	logger := jsonlog.New(os.Stdout, jsonlog.LEVEL_INFO)

	db, err := database.OpenDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	if err := db.AutoMigrate(&data.Movie{}, &data.Review{}); err != nil {
		logger.PrintFatal(err, nil)
	}
	defer func() {
		conn, _ := db.DB()
		conn.Close()
	}()
	logger.PrintInfo("database connection pool established", nil)

	app := &Application{config: cfg, logger: logger}
	appModels := models.NewModels(db)
	ginHandler := handler.NewHandler(appModels, logger)

	movieConsumer, err := kafka.NewConsumer(nil, address, moviesTopic, moviesGroup)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	reviewConsumer, err := kafka.NewConsumer(nil, address, reviewsTopic, reviewsGroup)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	go movieConsumer.StartWithFunc(ginHandler.HandleMovieMessage)
	go reviewConsumer.StartWithFunc(ginHandler.HandleReviewMessage)

	if err := app.serve(ginHandler.Routes()); err != nil {
		logger.PrintFatal(err, nil)
	}
	if err := movieConsumer.Stop(); err != nil {
		app.logger.PrintError(err, nil)
	}
	if err := reviewConsumer.Stop(); err != nil {
		app.logger.PrintError(err, nil)
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
		app.logger.PrintInfo("caught signal, shutting down", map[string]string{"signal": s.String()})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(ctx)
	}()

	app.logger.PrintInfo("starting Gin server", map[string]string{
		"env":  app.config.Env,
		"addr": srv.Addr,
	})

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownErr; err != nil {
		return err
	}

	app.logger.PrintInfo("server stopped", map[string]string{"addr": srv.Addr})
	return nil
}
