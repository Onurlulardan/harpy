package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

func writeAuthNoContent(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set(requestmeta.CorrelationHeader, authHTTPCorrelationID(r).String())
	w.WriteHeader(http.StatusNoContent)
}
