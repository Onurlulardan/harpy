package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestHostTokenExchangeAuthenticatesSubjectAndGRPCBearer(t *testing.T) {
	now := time.Date(2026, 7, 23, 3, 0, 0, 0, time.UTC)
	workloads, privateKey := fixtureWorkloadRegistry()
	authority := newTokenAuthority(
		"host", fixtureKey, workloads, func() time.Time { return now },
	)
	subject := signFixtureSubject(t, privateKey, now)
	values := validTokenValues()
	values.Set("audience", hostAudience)
	values.Set("subject_token", subject)
	request := tokenRequest(values)
	response := httptest.NewRecorder()
	tokenExchange(response, request, authority)
	if response.Code != http.StatusOK {
		t.Fatalf("exchange status=%d body=%s", response.Code, response.Body.String())
	}
	var token tokenResponse
	if err := json.Unmarshal(response.Body.Bytes(), &token); err != nil {
		t.Fatal(err)
	}
	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("authorization", "Bearer "+token.AccessToken))
	called := false
	_, err := authority.authorize(ctx, struct{}{}, &grpc.UnaryServerInfo{},
		func(context.Context, any) (any, error) {
			called = true
			return struct{}{}, nil
		})
	if err != nil || !called {
		t.Fatalf("bearer rejected: called=%t err=%v", called, err)
	}
}

func TestHostTokenExchangeRejectsTamperedSubject(t *testing.T) {
	workloads, _ := fixtureWorkloadRegistry()
	authority := newTokenAuthority("host", fixtureKey, workloads, time.Now)
	values := validTokenValues()
	values.Set("audience", hostAudience)
	values.Set("subject_token", "invalid")
	response := httptest.NewRecorder()
	tokenExchange(response, tokenRequest(values), authority)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func tokenRequest(values url.Values) *http.Request {
	request := httptest.NewRequest(
		http.MethodPost, "/v1/token", strings.NewReader(values.Encode()),
	)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return request
}
