package config

import (
	"net"
	"strings"
	"testing"
)

func TestAnalyticsTrustedProxiesAreCanonicalAndBounded(t *testing.T) {
	loaded, err := loadAPI(withEnvironment(validAPIEnvironment(),
		analyticsTrustedProxyKey+"=10.20.0.0/16,2001:db8::/48"))
	if err != nil {
		t.Fatal(err)
	}
	proxies := loaded.AnalyticsTrustedProxies()
	if proxies.Count() != 2 || !proxies.Contains(net.ParseIP("10.20.2.3")) ||
		!proxies.Contains(net.ParseIP("2001:db8::9")) ||
		proxies.Contains(net.ParseIP("203.0.113.1")) {
		t.Fatalf("unexpected trusted proxy set")
	}
	for _, value := range []string{"0.0.0.0/0", "10.0.0.1/8", "10.0.0.0/8,10.1.0.0/16"} {
		_, err = loadAPI(withEnvironment(validAPIEnvironment(),
			analyticsTrustedProxyKey+"="+value))
		if err == nil || !strings.Contains(err.Error(), analyticsTrustedProxyKey) {
			t.Fatalf("trusted proxy %q error = %v", value, err)
		}
	}
}
