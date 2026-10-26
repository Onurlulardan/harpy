package httpserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func NewPublic(
	address string,
	handler http.Handler,
	httpConfig config.HTTP,
	errorLog *log.Logger,
) (*http.Server, error) {
	if address == "" {
		return nil, fmt.Errorf("public HTTP address is required")
	}
	if handler == nil {
		return nil, fmt.Errorf("public HTTP handler is required")
	}
	if errorLog == nil {
		return nil, fmt.Errorf("public HTTP error logger is required")
	}
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadTimeout:       0,
		ReadHeaderTimeout: httpConfig.ReadHeaderTimeout(),
		WriteTimeout:      0,
		IdleTimeout:       httpConfig.IdleTimeout(),
		MaxHeaderBytes:    httpConfig.MaxHeaderBytes(),
		ErrorLog:          errorLog,
	}, nil
}
