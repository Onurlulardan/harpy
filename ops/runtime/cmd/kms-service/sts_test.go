package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestTokenExchangeReturnsBoundedBearerToken(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/v1/token", strings.NewReader(
		validTokenValues().Encode(),
	))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	tokenExchange(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("status = %d, cache = %q", response.Code, response.Header().Get("Cache-Control"))
	}
	var value tokenResponse
	if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if value.AccessToken == "" || value.TokenType != "Bearer" ||
		value.IssuedTokenType != accessTokenType || value.ExpiresIn != 3600 {
		t.Fatalf("unexpected token response: %+v", value)
	}
}

func TestTokenExchangeRejectsInvalidContract(t *testing.T) {
	tests := map[string]func(*http.Request){
		"method":     func(request *http.Request) { request.Method = http.MethodGet },
		"media type": func(request *http.Request) { request.Header.Del("Content-Type") },
		"subject": func(request *http.Request) {
			request.Body = http.NoBody
		},
		"query": func(request *http.Request) { request.URL.RawQuery = "extra=value" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/v1/token", strings.NewReader(
				validTokenValues().Encode(),
			))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			mutate(request)
			response := httptest.NewRecorder()
			tokenExchange(response, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", response.Code)
			}
		})
	}
}

func validTokenValues() url.Values {
	return url.Values{
		"audience": {"//iam.googleapis.com/projects/123456789/locations/global/" +
			"workloadIdentityPools/local/providers/runtime"},
		"grant_type":           {tokenGrantType},
		"requested_token_type": {accessTokenType},
		"subject_token_type":   {jwtTokenType},
		"subject_token":        {"local-short-lived-token"},
		"scope":                {cloudKMSScope},
	}
}
