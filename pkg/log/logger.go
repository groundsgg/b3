package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type ctxKeyLogger struct{}

// BuildRootLogger returns a JSON slog.Logger writing to stdout.
func BuildRootLogger() *slog.Logger {
	options := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	invalidLogLevel := false

	switch strings.ToLower(os.Getenv("LOGGING_LEVEL")) {
	case "debug":
		options.Level = slog.LevelDebug
	case "info", "":
		options.Level = slog.LevelInfo
	case "warn":
		options.Level = slog.LevelWarn
	case "error":
		options.Level = slog.LevelError
	default:
		invalidLogLevel = true
	}

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, options),
	)

	if invalidLogLevel {
		logger.Warn("invalid logging level; using info", "level", os.Getenv("LOGGING_LEVEL"))
	}

	return logger
}

// WithLogger stores the logger in the context for downstream handlers.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger{}, logger)
}

// LoggerFromContext returns the logger stored in the context or slog.Default if missing.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKeyLogger{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
