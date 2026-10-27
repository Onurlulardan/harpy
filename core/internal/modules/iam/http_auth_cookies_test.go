package iam

import (
	"strings"
	"testing"
	"time"
)

func TestRefreshCookieNeverExceedsFamilyAbsoluteExpiry(t *testing.T) {
	now := time.Unix(2_000_000_000, 0).UTC()
	absolute := now.Add(time.Hour + 750*time.Millisecond)
	value, valid := refreshCookieValue(
		[]byte(strings.Repeat("a", 22)+"."+strings.Repeat("b", 43)),
		absolute,
		now,
	)
	if !valid || !strings.Contains(value, "Max-Age=3600") ||
		strings.Contains(value, "Domain=") || !strings.Contains(value, "HttpOnly") ||
		!strings.Contains(value, "Secure") || !strings.Contains(value, "SameSite=Lax") {
		t.Fatalf("unexpected refresh cookie: %s", value)
	}
}
