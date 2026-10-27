package main

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForTargetsRetriesHTTPUntilTwoHundredResponse(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 2 {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	config := checkConfig{httpTargets: []string{server.URL}, timeout: 2 * time.Second}
	client, err := newHTTPClient("", config.attemptTimeout())
	if err != nil {
		t.Fatalf("newHTTPClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.timeout)
	defer cancel()
	if err = waitForTargets(ctx, config, client); err != nil {
		t.Fatalf("waitForTargets: %v", err)
	}
	if attempts.Load() < 2 {
		t.Fatal("unhealthy HTTP target was not retried")
	}
}

func TestHTTPClientTrustsExplicitCAWithoutSkippingVerification(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	caFile := filepath.Join(t.TempDir(), "local-ca.crt")
	certificate := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE", Bytes: server.Certificate().Raw,
	})
	if err := os.WriteFile(caFile, certificate, 0o600); err != nil {
		t.Fatalf("write CA: %v", err)
	}
	client, err := newHTTPClient(caFile, time.Second)
	if err != nil {
		t.Fatalf("newHTTPClient: %v", err)
	}
	transport := client.Transport.(*http.Transport)
	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("TLS verification was disabled")
	}
	if err = probeHTTP(context.Background(), client, server.URL); err != nil {
		t.Fatalf("probeHTTPS: %v", err)
	}
}

func TestProbeHTTPRejectsRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, "/ready", http.StatusFound)
	}))
	defer server.Close()
	client, err := newHTTPClient("", time.Second)
	if err != nil {
		t.Fatalf("newHTTPClient: %v", err)
	}
	err = probeHTTP(context.Background(), client, server.URL)
	if err == nil || !strings.Contains(err.Error(), "status 302") {
		t.Fatalf("redirect result = %v", err)
	}
}
