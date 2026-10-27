package iam

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func TestWorkflowPublicationAuthorityRequiresTargetBoundFreshStepUp(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	definitionID := uuid.MustParse("018f0000-0000-7000-8000-000000000002")
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	hash := sha256.New()
	_, _ = hash.Write([]byte(
		"harpy/workflow/publication-authority/step-up/v1\x00",
	))
	_, _ = hash.Write(tenantID[:])
	_, _ = hash.Write(definitionID[:])
	purpose := hash.Sum(nil)
	profile, err := accesstoken.NewProfile(
		"http://localhost:8000", "http://localhost:5173", true,
	)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := accesstoken.NewClaims(profile, accesstoken.ClaimsInput{
		Subject: uuid.New(), Session: uuid.New(), Tenant: tenantID,
		Client: accesstoken.ClientOperatorConsole, IssuedAt: now,
		JTI: uuid.New(), TokenVersion: 1, AuthorizationVersion: 1,
		PlatformAuthorizationClock: 1, SecurityEpoch: [32]byte{1},
		AccessMode:         accesstoken.AccessModeOperator,
		AuthenticationTime: now.Add(-time.Minute),
		Assurance:          accesstoken.AssuranceStepUp,
		Methods:            []accesstoken.AuthMethod{accesstoken.AuthMethodPassword},
		StepUpPurposeHash:  purpose,
	})
	if err != nil {
		t.Fatal(err)
	}
	gotPurpose, authenticatedAt, password, webAuthn, totp, valid :=
		workflowPublicationAuthorityStepUp(
			claims, definitionID.String(), now,
		)
	var expected [sha256.Size]byte
	copy(expected[:], purpose)
	if !valid || gotPurpose != expected || !password || webAuthn || totp ||
		authenticatedAt != now.Add(-time.Minute) {
		t.Fatalf("step-up valid=%v methods=%v/%v/%v at=%v",
			valid, password, webAuthn, totp, authenticatedAt)
	}
	other := uuid.MustParse("018f0000-0000-7000-8000-000000000003")
	if _, _, _, _, _, valid = workflowPublicationAuthorityStepUp(
		claims, other.String(), now,
	); valid {
		t.Fatal("step-up must be bound to the exact workflow definition")
	}
	if _, _, _, _, _, valid = workflowPublicationAuthorityStepUp(
		claims, definitionID.String(), now.Add(11*time.Minute),
	); valid {
		t.Fatal("stale step-up must be rejected")
	}
}
