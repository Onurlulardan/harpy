package csrf

import "testing"

func TestOriginPolicyIsExactAndFailClosed(t *testing.T) {
	policy, err := NewOriginPolicy(
		[]string{"https://app.harpy.io", "https://portal.harpy.io:8443"}, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !policy.Allows("https://app.harpy.io") || policy.Allows("https://evil.test") {
		t.Fatal("origin allowlist mismatch")
	}
	invalid := [][]string{
		nil,
		{"*"},
		{"null"},
		{"http://app.harpy.io"},
		{"https://app.harpy.io/path"},
		{"https://app.harpy.io", "https://app.harpy.io"},
	}
	for _, origins := range invalid {
		if _, err := NewOriginPolicy(origins, false); err != ErrInvalidPolicy {
			t.Fatalf("invalid origins accepted: %#v, %v", origins, err)
		}
	}
	if _, err := NewOriginPolicy([]string{"http://127.0.0.1:5173"}, true); err != nil {
		t.Fatalf("explicit local origin rejected: %v", err)
	}
}

func TestCookieContractIsNarrowAndHostBound(t *testing.T) {
	if RefreshCookieName != "__Secure-harpy-refresh" || RefreshCookiePath != "/api/auth" ||
		CSRFCookieName != "__Host-harpy-csrf" || CSRFCookiePath != "/" ||
		HeaderName != "X-CSRF-Token" {
		t.Fatal("browser cookie contract drifted")
	}
}
