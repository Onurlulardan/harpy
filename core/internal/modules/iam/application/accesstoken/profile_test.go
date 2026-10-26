package accesstoken

import "testing"

func TestProfileRequiresCanonicalHTTPSOrExplicitLocalHTTP(t *testing.T) {
	valid := [][2]string{
		{"https://auth.harpy.io", "https://api.harpy.io"},
		{"https://example.test/auth", "https://example.test/api"},
	}
	for _, pair := range valid {
		if _, err := NewProfile(pair[0], pair[1], false); err != nil {
			t.Fatalf("valid profile rejected: %v", err)
		}
	}
	invalid := []string{
		"", " http://localhost:8000", "http://example.com", "https://user@example.com",
		"https://example.com/path?query=1", "https://example.com/#fragment",
	}
	for _, value := range invalid {
		if _, err := NewProfile(value, "https://api.harpy.io", false); err != ErrInvalidProfile {
			t.Fatalf("invalid endpoint %q accepted: %v", value, err)
		}
	}
	if _, err := NewProfile("http://localhost:8000/auth", "http://127.0.0.1:8000/api", true); err != nil {
		t.Fatalf("explicit local profile rejected: %v", err)
	}
}
