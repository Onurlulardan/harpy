package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestClaimInitialPasswordRejectsUnboundStore(t *testing.T) {
	if _, err := (*store)(nil).ClaimInitialPassword(
		context.Background(), authentication.CredentialScope{},
		authentication.InitialPasswordClaim{},
	); err == nil {
		t.Fatal("initial password claim accepted an unbound store")
	}
}

func TestCanonicalInitialPasswordClaimedUsesUTC(t *testing.T) {
	location := time.FixedZone("local-test", 3*60*60)
	issuedAt := time.Date(2026, 7, 27, 14, 0, 0, 0, location)
	result := canonicalInitialPasswordClaimed(authentication.InitialPasswordClaimed{
		Revision: 1, IssuedAt: issuedAt, ExpiresAt: issuedAt.Add(10 * time.Minute),
	})
	if result.IssuedAt.Location() != time.UTC ||
		result.ExpiresAt.Location() != time.UTC ||
		!result.IssuedAt.Equal(issuedAt) ||
		!result.ExpiresAt.Equal(issuedAt.Add(10*time.Minute)) {
		t.Fatalf("claim timestamps are not canonical UTC: %#v", result)
	}
}
