package routers

import (
	"context"
	"net/http"

	"github.com/groundsgg/b3/internal/web/request"
)

type ctxReq struct{}

type Router struct {
	mux *http.ServeMux
}

func (r *Router) getCTXReq(ctx context.Context) *request.Request {
	if req, ok := ctx.Value(ctxReq{}).(*request.Request); ok {
		return req
	}

	return nil
}

func (r *Router) Handle(pattern string, handler func(*request.Request)) {
	r.mux.HandleFunc(pattern, func(ow http.ResponseWriter, or *http.Request) {
		if req := r.getCTXReq(or.Context()); req != nil {
			handler(req)
		}
	})
}

func (r *Router) Serve(req *request.Request) {
	if ctxR := r.getCTXReq(req.OriginalRequest.Context()); ctxR == nil {
		newR := req.OriginalRequest.WithContext(
			context.WithValue(req.OriginalRequest.Context(), ctxReq{}, req),
		)
		req.OriginalRequest = newR
	}
	r.mux.ServeHTTP(req.OriginalWriter, req.OriginalRequest)
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}
