package smtp

import (
	"net"
	"testing"
)

func TestSafeProviderIPRejectsInternalAndMetadataEndpoints(t *testing.T) {
	unsafe := []string{"127.0.0.1", "10.0.0.1", "169.254.169.254", "::1", "fc00::1"}
	for _, value := range unsafe {
		if safeProviderIP(net.ParseIP(value)) {
			t.Fatalf("safeProviderIP(%q) = true", value)
		}
	}
	if !safeProviderIP(net.ParseIP("203.0.113.10")) {
		t.Fatal("public documentation address rejected")
	}
}
