package httpinput

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type decodeFixture struct {
	Name  string      `json:"name"`
	Count interface{} `json:"count"`
}

func TestDecodeJSONAcceptsOneStrictObject(t *testing.T) {
	request := jsonRequest(`{"name":"forge","count":2}`)
	var target decodeFixture
	if err := DecodeJSON(httptest.NewRecorder(), request, 1024, &target); err != nil {
		t.Fatal(err)
	}
	if target.Name != "forge" || target.Count == nil {
		t.Fatalf("unexpected decoded value: %#v", target)
	}
}

func TestDecodeJSONRejectsAmbiguousOrUnknownDocuments(t *testing.T) {
	tests := []string{
		``, `null`, `[]`, `{"name":"a","name":"b","count":1}`,
		`{"name":"a","count":1}{}`, `{"name":"a","count":1,"extra":true}`,
		`{"name":"a","count":9007199254740992}`,
	}
	for _, document := range tests {
		request := jsonRequest(document)
		var target decodeFixture
		if err := DecodeJSON(httptest.NewRecorder(), request, 1024, &target); !errors.Is(err, ErrInvalid) {
			t.Fatalf("document %q returned %v", document, err)
		}
	}
}

func TestDecodeJSONRejectsUnsupportedMediaTypes(t *testing.T) {
	tests := []func(*http.Request){
		func(request *http.Request) { request.Header.Del("Content-Type") },
		func(request *http.Request) { request.Header.Set("Content-Type", "text/json") },
		func(request *http.Request) { request.Header.Set("Content-Type", "application/json; charset=latin1") },
		func(request *http.Request) { request.Header.Set("Content-Encoding", "gzip") },
		func(request *http.Request) { request.Header.Add("Content-Type", "application/json") },
	}
	for index, mutate := range tests {
		request := jsonRequest(`{"name":"a","count":1}`)
		mutate(request)
		var target decodeFixture
		if err := DecodeJSON(httptest.NewRecorder(), request, 1024, &target); !errors.Is(err, ErrMediaTypeUnsupported) {
			t.Fatalf("case %d returned %v", index, err)
		}
	}
}

func TestDecodeJSONEnforcesDeclaredAndStreamingLimits(t *testing.T) {
	declared := jsonRequest(`{"name":"a","count":1}`)
	declared.ContentLength = 2048
	var target decodeFixture
	if err := DecodeJSON(httptest.NewRecorder(), declared, 32, &target); !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("declared limit returned %v", err)
	}
	streaming := jsonRequest(`{"name":"` + strings.Repeat("a", 64) + `","count":1}`)
	streaming.ContentLength = -1
	if err := DecodeJSON(httptest.NewRecorder(), streaming, 32, &target); !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("streaming limit returned %v", err)
	}
}

func jsonRequest(document string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/resource", strings.NewReader(document))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	return request
}
