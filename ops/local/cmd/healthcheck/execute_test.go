package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteChecksHTTPAndTCPTogether(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	if err = execute([]string{
		"-http", server.URL,
		"-tcp", listener.Addr().String(),
		"-timeout", "1s",
	}); err != nil {
		t.Fatalf("execute: %v", err)
	}
}
