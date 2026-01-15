// SPDX-License-Identifier: AGPL-3.0-or-later
package middleware

import (
	"cmp"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/groundsgg/b3/internal/web/request"
)

type statusCodeWriter struct {
	w    http.ResponseWriter
	code int
}

// Header returns the header map from the wrapped ResponseWriter.
func (sw *statusCodeWriter) Header() http.Header {
	return sw.w.Header()
}

// Write forwards the write to the wrapped ResponseWriter.
func (sw *statusCodeWriter) Write(b []byte) (int, error) {
	return sw.w.Write(b)
}

// WriteHeader records the status code before sending the response.
func (sw *statusCodeWriter) WriteHeader(statusCode int) {
	sw.code = statusCode
	sw.w.WriteHeader(statusCode)
}

// HTTPLogging attaches request IDs and structured logs around handlers.
func HTTPLogging(parentLogger *slog.Logger) Middleware {
	return func(next Handler) Handler {
		return func(req *request.Request) {
			start := time.Now()
			reqID := uuid.NewString()
			w := req.OriginalWriter
			r := req.OriginalRequest

			logger := parentLogger.With(
				"req_id", reqID,
			)
			w.Header().Set("X-Request-ID", reqID)
			sw := &statusCodeWriter{w: w}

			req.OriginalWriter = sw
			req.Logger = logger
			req.ID = reqID

			next(req)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", cmp.Or(
					r.Header.Get("X-Forwarded-For"),
					r.Header.Get("X-Real-Ip"),
					r.RemoteAddr,
				),
				"code", sw.code,
			)

			logger.Debug("request info",
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}
	}
}
