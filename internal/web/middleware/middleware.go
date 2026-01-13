package middleware

import "net/http"

type Middleware func(next http.Handler) http.Handler

func Combine(middlewares ...Middleware) Middleware {
	if len(middlewares) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		stack := middlewares[len(middlewares)-1](next)
		for i := len(middlewares) - 1; i > 0; i-- {
			stack = middlewares[i-1](stack)
		}
		return stack
	}
}
