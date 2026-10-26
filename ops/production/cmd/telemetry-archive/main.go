package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "telemetry-archive:", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := loadConfig(os.LookupEnv)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr: config.listen,
		Handler: archiveHandler{
			directory: config.directory,
			clock:     time.Now,
		},
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 << 10,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()
	stopped := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(shutdown)
		case <-stopped:
		}
	}()
	err = server.ListenAndServeTLS(config.cert, config.key)
	close(stopped)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
