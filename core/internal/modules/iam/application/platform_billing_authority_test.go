package application

import (
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	"github.com/google/uuid"
)

func platformBillingAuthorityFixture(
	t *testing.T, capability string,
) (uuid.UUID, PlatformBillingAuthority) {
	t.Helper()
	actorID := uuid.New()
	authority, err := NewPlatformBillingAuthority(
		platformdomain.Decision{
			Allowed: true, Capability: capability, DeploymentMode: "saas",
			SubjectID: platformdomain.ID(actorID), PlatformClock: 7,
			AuthorizationVersion: 5, Reason: "granted",
		},
		actorID, capability, 5,
	)
	if err != nil {
		t.Fatalf("build platform Billing authority: %v", err)
	}
	return actorID, authority
}

func TestPlatformBillingAuthorityRejectsTenantAndBooleanStyleEvidence(t *testing.T) {
	actorID := uuid.New()
	for name, decision := range map[string]platformdomain.Decision{
		"tenant_all": {
			Allowed: true, Capability: "tenant_all", DeploymentMode: "saas",
			SubjectID: platformdomain.ID(actorID), PlatformClock: 7,
			AuthorizationVersion: 5, Reason: "granted",
		},
		"boolean_without_subject": {
			Allowed: true, Capability: "platform.billing.catalog.read",
			DeploymentMode: "saas", PlatformClock: 7,
			AuthorizationVersion: 5, Reason: "granted",
		},
		"onpremise_mode": {
			Allowed: true, Capability: "platform.billing.catalog.read",
			DeploymentMode: "onpremise", SubjectID: platformdomain.ID(actorID),
			PlatformClock: 7, AuthorizationVersion: 5, Reason: "granted",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewPlatformBillingAuthority(
				decision, actorID, "platform.billing.catalog.read", 5,
			); err == nil {
				t.Fatal("non-platform evidence minted global Billing authority")
			}
		})
	}
}
