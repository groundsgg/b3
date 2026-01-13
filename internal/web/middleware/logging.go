package middleware

import (
	"cmp"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/groundsgg/b3/pkg/log"
)

type ctxReqID struct{}

type statusCodeWriter struct {
	w    http.ResponseWriter
	code int
}

func (sw *statusCodeWriter) Header() http.Header {
	return sw.w.Header()
}

func (sw *statusCodeWriter) Write(b []byte) (int, error) {
	return sw.w.Write(b)
}

func (sw *statusCodeWriter) WriteHeader(statusCode int) {
	sw.code = statusCode
	sw.w.WriteHeader(statusCode)
}

func HTTPLogging(parentLogger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := uuid.NewString()

			logger := parentLogger.With(
				"req_id", reqID,
			)

			w.Header().Set("X-Request-ID", reqID)

			ctx := log.WithLogger(r.Context(), logger)
			ctx = context.WithValue(ctx, ctxReqID{}, reqID)
			sw := &statusCodeWriter{w: w}

			next.ServeHTTP(sw, r.WithContext(ctx))

			logger.Info("request received",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", cmp.Or(
					r.Header.Get("X-Forwarded-For"),
					r.Header.Get("X-Real-Ip"),
					r.RemoteAddr,
				),
				"code", sw.code,
			)

			logger.Debug("request completed",
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
