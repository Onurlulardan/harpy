package httpserver

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestBoundServerServesAndShutsDown(t *testing.T) {
	server, err := NewManagement(
		"127.0.0.1:0",
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		log.New(io.Discard, "", 0),
	)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := Bind(context.Background(), server, net.ListenConfig{})
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- bound.Serve() }()
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get("http://" + bound.Address().String() + "/livez")
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
	shutdownContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := bound.Shutdown(shutdownContext); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-served:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP server did not stop")
	}
}

func TestBindRejectsListenerCollision(t *testing.T) {
	firstServer, err := NewManagement(
		"127.0.0.1:0",
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		log.New(io.Discard, "", 0),
	)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Bind(context.Background(), firstServer, net.ListenConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	secondServer, err := NewManagement(
		first.Address().String(),
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		log.New(io.Discard, "", 0),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Bind(context.Background(), secondServer, net.ListenConfig{}); err == nil {
		t.Fatal("expected listener collision")
	}
}

func TestCloseReleasesListenerBeforeServe(t *testing.T) {
	server, err := NewManagement(
		"127.0.0.1:0",
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		log.New(io.Discard, "", 0),
	)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := Bind(context.Background(), server, net.ListenConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := bound.Close(); err != nil {
		t.Fatal(err)
	}
	if err := bound.Close(); err != nil {
		t.Fatalf("second close must be idempotent: %v", err)
	}
}
