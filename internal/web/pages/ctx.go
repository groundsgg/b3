package pages

import (
	"context"
	"net/http"
)

type PageData struct {
	Title string
	Data  any
}

type Pages interface {
	Render(w http.ResponseWriter, name string, data PageData) error
}

type ctxKeyPages struct{}

func WithPages(ctx context.Context, pages Pages) context.Context {
	return context.WithValue(ctx, ctxKeyPages{}, pages)
}

func PagesFromContext(ctx context.Context) Pages {
	if l, ok := ctx.Value(ctxKeyPages{}).(Pages); ok {
		return l
	}
	return nil
}
