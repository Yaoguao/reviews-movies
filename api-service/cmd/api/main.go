package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reviews-movies/api-service/config"
	"reviews-movies/api-service/interanal/handler"
	"reviews-movies/api-service/interanal/jsonlog"
	"reviews-movies/api-service/interanal/kafka"
	"strings"
	"syscall"
	"time"
)

type Application struct {
	config config.Config
	logger *jsonlog.Logger
}

func main() {
	var cfg config.Config

	flag.IntVar(&cfg.Port, "port", 8082, "API server port")
	flag.StringVar(&cfg.Env, "env", "development",
		"Environment (development|staging|production)")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LEVEL_INFO)

	app := &Application{
		config: cfg,
		logger: logger,
	}

	addresses := []string{
		os.Getenv("BROKER1_HOST"),
		os.Getenv("BROKER2_HOST"),
		os.Getenv("BROKER3_HOST"),
	}

	logger.PrintInfo(strings.Join(addresses, ", "), nil)

	producer, err := kafka.NewProducer(addresses)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	newHandler := handler.NewHandler(producer)
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

		app.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.PrintInfo("starting server", map[string]string{
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

	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
