package middleware

import (
	"github.com/groundsgg/b3/internal/web/request"
)

type Handler func(req *request.Request)

type Middleware func(next Handler) Handler

// Combine chains middlewares into a single middleware in the given order.
func Combine(middlewares ...Middleware) Middleware {
	if len(middlewares) == 0 {
		return func(next Handler) Handler {
			return next
		}
	}

	return func(next Handler) Handler {
		stack := middlewares[len(middlewares)-1](next)
		for i := len(middlewares) - 1; i > 0; i-- {
			stack = middlewares[i-1](stack)
		}
		return stack
	}
}
