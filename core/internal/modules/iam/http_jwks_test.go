package iam

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func TestJWKSHandlerGETHEADAndConditionalRead(t *testing.T) {
	handler := newJWKSFixture(t)
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil))
	if get.Code != http.StatusOK || get.Body.Len() == 0 ||
		get.Header().Get("Content-Type") != "application/jwk-set+json" ||
		get.Header().Get("Cache-Control") != "public, max-age=300, must-revalidate" ||
		get.Header().Get("ETag") == "" || get.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("unexpected GET response: %d %v %q", get.Code, get.Header(), get.Body.String())
	}
	head := httptest.NewRecorder()
	handler.ServeHTTP(head, httptest.NewRequest(http.MethodHead, "/.well-known/jwks.json", nil))
	if head.Code != http.StatusOK || head.Body.Len() != 0 ||
		head.Header().Get("Content-Length") != get.Header().Get("Content-Length") {
		t.Fatalf("unexpected HEAD response: %d %v %q", head.Code, head.Header(), head.Body.String())
	}
	request := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	request.Header.Set("If-None-Match", "W/"+get.Header().Get("ETag"))
	conditional := httptest.NewRecorder()
	handler.ServeHTTP(conditional, request)
	if conditional.Code != http.StatusNotModified || conditional.Body.Len() != 0 {
		t.Fatalf("unexpected conditional response: %d %q", conditional.Code, conditional.Body.String())
	}
}

func TestJWKSHandlerRejectsMutationMethods(t *testing.T) {
	response := httptest.NewRecorder()
	newJWKSFixture(t).ServeHTTP(
		response, httptest.NewRequest(http.MethodPost, "/.well-known/jwks.json", nil),
	)
	if response.Code != http.StatusMethodNotAllowed ||
		response.Header().Get("Allow") != "GET, HEAD" ||
		response.Header().Get("Cache-Control") != "no-store" ||
		response.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("unexpected method response: %d %v", response.Code, response.Header())
	}
}

func newJWKSFixture(t *testing.T) *JWKSHandler {
	t.Helper()
	keyID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 12))
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{2}, ed25519.SeedSize))
	keys, err := accesstoken.NewPublicKeySet(map[string]ed25519.PublicKey{
		keyID: private.Public().(ed25519.PublicKey),
	})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewJWKSHandler(keys, func(err error) { t.Errorf("handler error: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	return handler
}
