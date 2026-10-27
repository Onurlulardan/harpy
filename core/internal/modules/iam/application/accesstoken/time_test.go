package accesstoken

import (
	"context"
	"testing"
	"time"
)

func TestVerifierUsesExactLifetimeAndClockSkew(t *testing.T) {
	fixture := newTokenFixture(t)
	signer := fixture.signer(t)
	defer signer.Close()
	token, err := signer.Sign(fixture.claims(t))
	if err != nil {
		t.Fatal(err)
	}
	verifier := fixture.verifier(t)
	tests := []struct {
		name   string
		now    time.Time
		accept bool
	}{
		{name: "exact early skew", now: fixture.now.Add(-ClockSkew), accept: true},
		{name: "too early", now: fixture.now.Add(-ClockSkew - time.Second)},
		{name: "last expiry skew second", now: fixture.now.Add(Lifetime + ClockSkew - time.Second), accept: true},
		{name: "expiry boundary", now: fixture.now.Add(Lifetime + ClockSkew)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := verifier.Verify(context.Background(), token, test.now)
			if test.accept && err != nil {
				t.Fatalf("expected acceptance, got %v", err)
			}
			if !test.accept && err != ErrInvalidToken {
				t.Fatalf("expected invalid token, got %v", err)
			}
		})
	}
}
