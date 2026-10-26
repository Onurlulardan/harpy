package httpwire

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONSetsCanonicalHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	err := WriteJSON(recorder, http.StatusOK, "request-1", map[string]string{"status": "ok"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" ||
		recorder.Header().Get("Cache-Control") != "no-store" ||
		recorder.Header().Get("X-Content-Type-Options") != "nosniff" ||
		recorder.Header().Get("X-Correlation-ID") != "request-1" {
		t.Fatalf("canonical headers missing: %#v", recorder.Header())
	}
	if recorder.Header().Get("Content-Length") == "" || recorder.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected response body: %q", recorder.Body.String())
	}
}

func TestWriteJSONHeadUsesGETContentLengthWithoutBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	if err := WriteJSON(recorder, http.StatusOK, "request-1", map[string]string{"status": "ok"}, true); err != nil {
		t.Fatal(err)
	}
	if recorder.Body.Len() != 0 || recorder.Header().Get("Content-Length") != "15" {
		t.Fatalf("unexpected HEAD response: length=%s body=%q", recorder.Header().Get("Content-Length"), recorder.Body.String())
	}
}

func TestWriteJSONMarshalsBeforeCommittingHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	if err := WriteJSON(recorder, http.StatusOK, "request-1", make(chan int), false); err == nil {
		t.Fatal("expected marshal failure")
	}
	if recorder.Code != http.StatusOK || len(recorder.Header()) != 0 || recorder.Body.Len() != 0 {
		t.Fatalf("response committed before marshal completed")
	}
}
