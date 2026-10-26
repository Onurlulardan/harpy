package api

import (
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestAnalyticsClientIPRejectsUntrustedForwarding(t *testing.T) {
	resolver := newAnalyticsClientIPResolver(config.AnalyticsTrustedProxies{})
	request := httptest.NewRequest("GET", "https://example.test/api/published", nil)
	request.RemoteAddr = "203.0.113.8:443"
	request.Header.Set("X-Forwarded-For", "198.51.100.2")
	if _, err := resolver(request); err == nil {
		t.Fatal("untrusted forwarding was accepted")
	}
	request.Header.Del("X-Forwarded-For")
	value, err := resolver(request)
	if err != nil || value != "203.0.113.8" {
		t.Fatalf("direct peer = %q/%v", value, err)
	}
}
