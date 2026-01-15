package core

import (
	"net/http"

	"github.com/groundsgg/b3/internal/web/request"
)

func A(req *request.Request) {
	req.OriginalWriter.WriteHeader(http.StatusOK)
	err := req.PrintOnly("pages.a", "AAA", nil)
	if err != nil {
		req.Logger.Error("failed to parse template",
			"template", "pages.a",
			"err", err,
		)
	}
}

func B(req *request.Request) {
	req.OriginalWriter.WriteHeader(http.StatusOK)
	err := req.PrintOnly("pages.b", "BBB", nil)
	if err != nil {
		req.Logger.Error("failed to parse template",
			"template", "pages.b",
			"err", err,
		)
	}
}
