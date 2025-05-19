package logger

import (
	"io"
	"log/slog"
)

var Logger *slog.Logger

func InitLogger(env string, w io.Writer) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{}
	switch env {
	case "development":
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(w, opts)
	default:
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(w, opts)
	}

	Logger = slog.New(handler)

	return Logger
}
