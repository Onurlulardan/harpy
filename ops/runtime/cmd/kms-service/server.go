package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"time"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
)

const shutdownTimeout = 5 * time.Second

func serve(ctx context.Context, config runtimeConfig) error {
	authority := newTokenAuthority(
		config.profile, config.tokenKey, config.workloads, time.Now,
	)
	grpcOptions := []grpc.ServerOption{grpc.MaxRecvMsgSize(256 * 1024)}
	if config.profile == "host" {
		grpcOptions = append(grpcOptions, grpc.UnaryInterceptor(authority.authorize))
	}
	grpcServer := grpc.NewServer(grpcOptions...)
	service := newKMSService(config.masterKey, config.profile)
	kmspb.RegisterKeyManagementServiceServer(grpcServer, service)
	server := &http.Server{
		Addr:              config.listenAddr,
		Handler:           newHTTPHandler(grpcServer, authority),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 * 1024,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"h2", "http/1.1"},
		},
	}
	if err := http2.ConfigureServer(server, &http2.Server{}); err != nil {
		return errors.New("local KMS HTTP/2 configuration failed")
	}
	stopped := make(chan struct{})
	go shutdownWhenCanceled(ctx, server, grpcServer, stopped)
	err := server.ListenAndServeTLS(config.certFile, config.tlsKeyFile)
	close(stopped)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func shutdownWhenCanceled(
	ctx context.Context,
	server *http.Server,
	grpcServer *grpc.Server,
	stopped <-chan struct{},
) {
	select {
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownContext)
		grpcServer.Stop()
	case <-stopped:
	}
}
