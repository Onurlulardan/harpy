package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

const maximumCABytes = 1024 * 1024

func newHTTPClient(caFile string, attemptTimeout time.Duration) (*http.Client, error) {
	roots, err := trustedRoots(caFile)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   attemptTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   attemptTimeout,
		ResponseHeaderTimeout: attemptTimeout,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			RootCAs:    roots,
		},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   attemptTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

func trustedRoots(caFile string) (*x509.CertPool, error) {
	roots, err := x509.SystemCertPool()
	if err != nil || roots == nil {
		roots = x509.NewCertPool()
	}
	if caFile == "" {
		return roots, nil
	}
	info, err := os.Stat(caFile)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maximumCABytes {
		return nil, errors.New("healthcheck -ca-file must name a bounded regular file")
	}
	content, err := os.ReadFile(caFile)
	if err != nil || !roots.AppendCertsFromPEM(content) {
		return nil, fmt.Errorf("healthcheck -ca-file does not contain a valid PEM CA")
	}
	return roots, nil
}
