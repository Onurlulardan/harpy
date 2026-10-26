package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	managementReadHeaderTimeout = 2 * time.Second
	managementReadTimeout       = 5 * time.Second
	managementWriteTimeout      = 5 * time.Second
	managementIdleTimeout       = 30 * time.Second
	managementMaxHeaderBytes    = 16 * 1024
)

func NewManagement(address string, handler http.Handler, errorLog *log.Logger) (*http.Server, error) {
	if address == "" {
		return nil, fmt.Errorf("management HTTP address is required")
	}
	if handler == nil {
		return nil, fmt.Errorf("management HTTP handler is required")
	}
	if errorLog == nil {
		return nil, fmt.Errorf("management HTTP error logger is required")
	}
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadTimeout:       managementReadTimeout,
		ReadHeaderTimeout: managementReadHeaderTimeout,
		WriteTimeout:      managementWriteTimeout,
		IdleTimeout:       managementIdleTimeout,
		MaxHeaderBytes:    managementMaxHeaderBytes,
		ErrorLog:          errorLog,
	}, nil
}
