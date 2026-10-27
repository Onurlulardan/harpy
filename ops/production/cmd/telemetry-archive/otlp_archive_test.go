package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHandlerArchivesExactOTLPBody(t *testing.T) {
	directory := t.TempDir()
	handler := archiveHandler{
		directory: directory,
		clock: func() time.Time {
			return time.Date(2026, 7, 23, 3, 0, 0, 0, time.UTC)
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/traces", http.NoBody)
	request.Body = ioBody("otlp")
	request.Header.Set("Content-Type", "application/x-protobuf")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 1 {
		t.Fatalf("archive entries=%d err=%v", len(entries), err)
	}
	content, err := os.ReadFile(directory + string(os.PathSeparator) + entries[0].Name())
	if err != nil || string(content) != "otlp" {
		t.Fatalf("archive=%q err=%v", content, err)
	}
}

func TestHandlerRejectsUnsupportedSurface(t *testing.T) {
	handler := archiveHandler{directory: t.TempDir(), clock: time.Now}
	for _, test := range []struct {
		path        string
		contentType string
		status      int
	}{
		{"/v1/traces?x=1", "application/x-protobuf", http.StatusNotFound},
		{"/v1/unknown", "application/x-protobuf", http.StatusNotFound},
		{"/v1/traces", "text/plain", http.StatusUnsupportedMediaType},
	} {
		request := httptest.NewRequest(http.MethodPost, test.path, ioBody("x"))
		request.Header.Set("Content-Type", test.contentType)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.status {
			t.Fatalf("%s status=%d want=%d", test.path, response.Code, test.status)
		}
	}
}

func ioBody(value string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(value))
}
