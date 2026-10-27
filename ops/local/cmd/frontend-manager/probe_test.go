package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFrontendReadyRequiresApplicationRoot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`<html><div id="root"></div></html>`))
	}))
	defer server.Close()
	if !frontendReady(context.Background(), server.URL) {
		t.Fatal("expected frontend readiness")
	}
}

func TestFrontendReadyRejectsUnrelatedHTTPServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte("not the frontend"))
	}))
	defer server.Close()
	if frontendReady(context.Background(), server.URL) {
		t.Fatal("unexpected readiness")
	}
}
