package middleware

import (
	"io"

	"github.com/groundsgg/b3/internal/web/request"
)

func drainAndCloseBody(body io.ReadCloser) {
	if body == nil {
		return
	}

	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}

// DrainBody drains and closes the request body after the handler finishes.
func DrainBody() Middleware {
	return func(next Handler) Handler {
		return func(req *request.Request) {
			defer drainAndCloseBody(req.OriginalRequest.Body)
			next(req)
		}
	}
}
