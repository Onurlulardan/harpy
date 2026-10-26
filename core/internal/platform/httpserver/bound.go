package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
)

type Bound struct {
	server   *http.Server
	listener net.Listener
}

func Bind(ctx context.Context, server *http.Server, listenConfig net.ListenConfig) (*Bound, error) {
	if server == nil {
		return nil, fmt.Errorf("HTTP server is required")
	}
	listener, err := listenConfig.Listen(ctx, "tcp", server.Addr)
	if err != nil {
		return nil, fmt.Errorf("bind HTTP listener: %w", err)
	}
	return &Bound{server: server, listener: listener}, nil
}

func (bound *Bound) Serve() error {
	err := bound.server.Serve(bound.listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("serve HTTP listener: %w", err)
	}
	return nil
}

func (bound *Bound) Shutdown(ctx context.Context) error {
	if err := bound.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown HTTP listener: %w", err)
	}
	if err := bound.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("close shutdown listener: %w", err)
	}
	return nil
}

func (bound *Bound) Close() error {
	if err := bound.server.Close(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("close HTTP listener: %w", err)
	}
	if err := bound.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("close bound listener: %w", err)
	}
	return nil
}

func (bound *Bound) Address() net.Addr {
	return bound.listener.Addr()
}
