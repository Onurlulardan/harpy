package iam

import (
	"context"
	"testing"
)

func TestAuthHTTPPeerCanonicalizesSocketAddress(t *testing.T) {
	tests := map[string]string{
		"203.0.113.9:443":       "203.0.113.9",
		"[2001:db8::1]:8443":    "2001:db8::1",
		"[::ffff:192.0.2.8]:80": "192.0.2.8",
		"192.0.2.7":             "192.0.2.7",
	}
	for remoteAddress, expected := range tests {
		ctx := withAuthHTTPPeer(context.Background(), remoteAddress)
		actual, valid := authHTTPPeerIP(ctx)
		if !valid || actual != expected {
			t.Fatalf("peer %q = %q/%t, want %q", remoteAddress, actual, valid, expected)
		}
	}
}

func TestAuthHTTPPeerRejectsMissingOrAmbiguousAddress(t *testing.T) {
	for _, remoteAddress := range []string{"", "proxy.internal:443", "0.0.0.0:1", " address "} {
		ctx := withAuthHTTPPeer(context.Background(), remoteAddress)
		if value, valid := authHTTPPeerIP(ctx); valid || value != "" {
			t.Fatalf("invalid peer %q accepted as %q", remoteAddress, value)
		}
	}
}
