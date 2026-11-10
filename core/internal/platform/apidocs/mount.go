package apidocs

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/contracts"
)

const specPrefix = "/docs/spec/"

func Mount(next http.Handler) (http.Handler, error) {
	if next == nil {
		return nil, fmt.Errorf("public HTTP handler is required")
	}
	if _, err := contracts.Files.Open("http/core-api-v1.openapi.yaml"); err != nil {
		return nil, fmt.Errorf("API reference contract is missing: %w", err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /docs", referenceHandler{})
	mux.Handle("GET /docs/{$}", http.RedirectHandler("/docs", http.StatusPermanentRedirect))
	mux.Handle("GET "+specPrefix, specHandler{})
	mux.Handle("/", next)
	return mux, nil
}
