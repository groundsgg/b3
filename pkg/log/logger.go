// SPDX-License-Identifier: AGPL-3.0-or-later
package log

import (
	"context"
	"log/slog"
	"os"
)

type ctxKeyLogger struct{}

// BuildRootLogger returns a JSON slog.Logger writing to stdout.
func BuildRootLogger(options *slog.HandlerOptions) *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, options),
	)
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
