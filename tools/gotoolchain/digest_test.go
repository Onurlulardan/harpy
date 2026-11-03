package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBookwormDigestReadsContentHeader(t *testing.T) {
	digest := "sha256:" + strings.Repeat("cd", 32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"registry-token"}`))
		case "/v2/library/golang/manifests/1.27.3-bookworm":
			if r.Header.Get("Authorization") != "Bearer registry-token" {
				t.Errorf("authorization = %q", r.Header.Get("Authorization"))
			}
			if !strings.Contains(r.Header.Get("Accept"), "application/vnd.oci.image.index.v1+json") {
				t.Errorf("accept = %q", r.Header.Get("Accept"))
			}
			w.Header().Set("Docker-Content-Digest", digest)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := registryClient{
		tokenURL:    server.URL + "/token",
		manifestURL: server.URL + "/v2/library/golang/manifests/",
		client:      server.Client(),
	}
	got, err := client.bookwormDigest(context.Background(), goRelease{major: 1, minor: 27, patch: 3})
	if err != nil {
		t.Fatal(err)
	}
	if got != digest {
		t.Fatalf("digest = %s", got)
	}
}

func TestBookwormDigestRejectsMissingTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"token":"registry-token"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	client := registryClient{
		tokenURL:    server.URL + "/token",
		manifestURL: server.URL + "/v2/library/golang/manifests/",
		client:      server.Client(),
	}
	_, err := client.bookwormDigest(context.Background(), goRelease{major: 1, minor: 27, patch: 3})
	if err == nil {
		t.Fatal("missing image tag was accepted")
	}
}
