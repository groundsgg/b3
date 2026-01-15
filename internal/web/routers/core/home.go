package core

import (
	"net/http"

	"github.com/groundsgg/b3/internal/web/request"
)

func Home(req *request.Request) {
	if req.OriginalRequest.URL.Path != "/" {
		req.PrintError(request.ErrorData{
			Code:    http.StatusNotFound,
			Message: "Page not found",
		})
		return
	}

	req.OriginalWriter.WriteHeader(http.StatusOK)
	err := req.PrintOnly("pages.home", "Welcome to B3", nil)
	if err != nil {
		req.Logger.Error("failed to parse template",
			"template", "pages.home",
			"err", err,
		)
	}
}
