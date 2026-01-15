package request

import (
	"net/http"

	"github.com/groundsgg/b3/internal/web/pages"
)

func Parse(f func(req *Request), pages pages.Pages) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &Request{
			OriginalWriter:  w,
			OriginalRequest: r,
			pages:           pages,
		}
		f(req)
	})
}
