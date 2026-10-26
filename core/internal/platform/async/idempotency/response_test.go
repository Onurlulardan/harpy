package idempotency

import (
	"bytes"
	"strings"
	"testing"
)

func TestResponseIsBoundedDefensiveAndDeterministic(t *testing.T) {
	t.Parallel()
	body := []byte(`{"id":"42"}`)
	headers := map[string]string{"Content-Type": "application/json", "ETag": `"v1"`}
	response, err := NewBodyResponse(201, headers, body)
	if err != nil {
		t.Fatal(err)
	}
	if err := response.Validate(); err != nil {
		t.Fatalf("new response did not validate: %v", err)
	}
	body[0] = 'x'
	headers["Content-Type"] = "text/plain"
	if !bytes.Equal(response.Body(), []byte(`{"id":"42"}`)) ||
		response.Headers()["content-type"] != "application/json" {
		t.Fatal("response retained caller-owned mutable input")
	}
	reordered, _ := NewBodyResponse(201, map[string]string{
		"etag": `"v1"`, "content-type": "application/json",
	}, []byte(`{"id":"42"}`))
	if response.SHA256() != reordered.SHA256() {
		t.Fatal("response hash depends on header map order or case")
	}
}

func TestResponseRejectsSecretsUnsafeHeadersAndOversize(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"Set-Cookie", "Authorization", "X-Api-Token", "Location"} {
		if _, err := NewBodyResponse(200, map[string]string{name: "secret"}, []byte("ok")); err == nil {
			t.Fatalf("unsafe header %s accepted", name)
		}
	}
	if _, err := NewBodyResponse(200, map[string]string{"ETag": "bad\r\nheader"}, nil); err == nil {
		t.Fatal("header injection accepted")
	}
	if _, err := NewBodyResponse(200, nil, make([]byte, maximumResponseBodyBytes+1)); err == nil {
		t.Fatal("oversized body accepted")
	}
	if _, err := NewReferenceResponse(200, nil, strings.Repeat("r", 2049)); err == nil {
		t.Fatal("oversized reference accepted")
	}
}

func TestEmptyBodyRemainsAStoredBodyInsteadOfNull(t *testing.T) {
	t.Parallel()
	response, err := NewBodyResponse(204, nil, nil)
	if err != nil || response.Validate() != nil || response.Body() == nil {
		t.Fatalf("empty body response = %#v, %v", response, err)
	}
}
