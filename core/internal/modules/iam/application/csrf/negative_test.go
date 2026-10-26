package csrf

import (
	"context"
	"testing"
	"time"
)

func TestCSRFVerifierRejectsSubstitutionAndTampering(t *testing.T) {
	fixture := newCSRFFixture(t)
	wire := fixture.issue(t)
	wrongSelector := fixture.selector
	wrongSelector[0] ^= 1
	tampered := append([]byte(nil), wire...)
	tampered[len(tampered)-1] ^= 1
	tests := []struct {
		name     string
		cookie   []byte
		header   []byte
		selector [16]byte
		origin   string
	}{
		{name: "missing", selector: fixture.selector, origin: fixture.origin},
		{name: "header mismatch", cookie: wire, header: tampered, selector: fixture.selector, origin: fixture.origin},
		{name: "mac tamper", cookie: tampered, header: tampered, selector: fixture.selector, origin: fixture.origin},
		{name: "selector substitution", cookie: wire, header: wire, selector: wrongSelector, origin: fixture.origin},
		{name: "origin substitution", cookie: wire, header: wire, selector: fixture.selector, origin: "https://evil.test"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := fixture.verifier(t).Verify(
				context.Background(), test.cookie, test.header,
				test.selector, test.origin, fixture.now,
			)
			if err != ErrInvalidCredential {
				t.Fatalf("expected invalid credential, got %v", err)
			}
		})
	}
}

func TestCSRFVerifierEnforcesLifetimeAndSkew(t *testing.T) {
	fixture := newCSRFFixture(t)
	wire := fixture.issue(t)
	tests := []struct {
		name   string
		now    time.Time
		accept bool
	}{
		{name: "early skew", now: fixture.now.Add(-ClockSkew), accept: true},
		{name: "too early", now: fixture.now.Add(-ClockSkew - time.Second)},
		{name: "last expiry second", now: fixture.now.Add(Lifetime + ClockSkew - time.Second), accept: true},
		{name: "expiry boundary", now: fixture.now.Add(Lifetime + ClockSkew)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := fixture.verifier(t).Verify(
				context.Background(), wire, wire, fixture.selector, fixture.origin, test.now,
			)
			if test.accept && err != nil {
				t.Fatal(err)
			}
			if !test.accept && err != ErrInvalidCredential {
				t.Fatalf("expected invalid credential, got %v", err)
			}
		})
	}
}

func TestCSRFVerifierRejectsInvalidExecutionContextAndTypedNilResolver(t *testing.T) {
	fixture := newCSRFFixture(t)
	wire := fixture.issue(t)
	verifier := fixture.verifier(t)
	if err := verifier.Verify(
		nil, wire, wire, fixture.selector, fixture.origin, fixture.now,
	); err != ErrInvalidCredential {
		t.Fatalf("nil context error = %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := verifier.Verify(
		cancelled, wire, wire, fixture.selector, fixture.origin, fixture.now,
	); err != ErrInvalidCredential {
		t.Fatalf("cancelled context error = %v", err)
	}
	var resolver KeyResolverFunc
	if _, err := NewVerifier(fixture.policy, resolver); err != ErrInvalidPolicy {
		t.Fatalf("typed nil resolver error = %v", err)
	}
}
