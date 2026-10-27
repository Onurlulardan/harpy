package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func TestOwnerProxyCompletesPrivateACL(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter,
		request *http.Request) {
		t.Fatal("ACL compatibility response must not depend on unsupported MinIO ACL authorization")
	}))
	defer upstream.Close()
	target, _ := url.Parse(upstream.URL)
	proxy, err := newOwnerProxy(target, "123456789012")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "http://s3/bucket?acl", nil)
	request.Header.Set("x-amz-expected-bucket-owner", "123456789012")
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, request)
	if response.Code != http.StatusOK ||
		!strings.Contains(response.Body.String(), "<ID>123456789012</ID>") ||
		!strings.Contains(response.Body.String(), string(types.TypeCanonicalUser)) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestOwnerProxyRejectsMismatch(t *testing.T) {
	target, _ := url.Parse("http://127.0.0.1:1")
	proxy, err := newOwnerProxy(target, "123456789012")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "http://s3/bucket?acl", nil)
	request.Header.Set("x-amz-expected-bucket-owner", "000000000000")
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, request)
	body, _ := io.ReadAll(response.Result().Body)
	if response.Code != http.StatusForbidden || !strings.Contains(string(body), "AccessDenied") {
		t.Fatalf("response = %d %s", response.Code, body)
	}
}

func TestOwnerValidation(t *testing.T) {
	for _, value := range []string{"", "123", "12345678901x", "1234567890123"} {
		if validOwner(value) {
			t.Fatalf("validOwner(%q) = true", value)
		}
	}
	if !validOwner("123456789012") {
		t.Fatal("12 digit owner rejected")
	}
}
