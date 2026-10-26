package main

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestWaitForTargetsAcceptsReachableTCPListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	config := checkConfig{
		tcpTargets: []string{listener.Addr().String()},
		timeout:    time.Second,
	}
	client, err := newHTTPClient("", config.attemptTimeout())
	if err != nil {
		t.Fatalf("newHTTPClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.timeout)
	defer cancel()
	if err = waitForTargets(ctx, config, client); err != nil {
		t.Fatalf("waitForTargets: %v", err)
	}
}

func TestWaitForTargetsReportsUnreachableTCPAtTimeout(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	target := listener.Addr().String()
	_ = listener.Close()
	config := checkConfig{tcpTargets: []string{target}, timeout: 250 * time.Millisecond}
	client, err := newHTTPClient("", config.attemptTimeout())
	if err != nil {
		t.Fatalf("newHTTPClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.timeout)
	defer cancel()
	err = waitForTargets(ctx, config, client)
	if err == nil || !containsAll(err.Error(), "timed out", "tcp "+target, "connection failed") {
		t.Fatalf("timeout result = %v", err)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
