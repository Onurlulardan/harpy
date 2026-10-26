package billing

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type subscriptionBoundaryClaimFixture struct {
	jobID, attemptID uuid.UUID
	tenantID         uuid.UUID
	kind             string
	revision         int
	reference        string
	digest           string
	leaseGeneration  int64
}

func (claim subscriptionBoundaryClaimFixture) JobID() uuid.UUID         { return claim.jobID }
func (claim subscriptionBoundaryClaimFixture) AttemptID() uuid.UUID     { return claim.attemptID }
func (claim subscriptionBoundaryClaimFixture) Kind() string             { return claim.kind }
func (claim subscriptionBoundaryClaimFixture) PayloadRevision() int     { return claim.revision }
func (claim subscriptionBoundaryClaimFixture) PayloadReference() string { return claim.reference }
func (claim subscriptionBoundaryClaimFixture) PayloadSHA256() string    { return claim.digest }
func (claim subscriptionBoundaryClaimFixture) LeaseGeneration() int64   { return claim.leaseGeneration }
func (claim subscriptionBoundaryClaimFixture) TenantID() (uuid.UUID, bool) {
	return claim.tenantID, claim.tenantID != uuid.Nil
}

func subscriptionBoundaryFixture(t *testing.T) (
	uuid.UUID,
	subscriptionapp.JobIntent,
	subscriptionBoundaryClaimFixture,
) {
	t.Helper()
	tenantID := uuid.New()
	intent := subscriptionapp.JobIntent{
		ID:             uuid.MustParse("019c6a7e-7c00-7000-8000-000000000040"),
		EffectID:       uuid.MustParse("019c6a7e-7c00-7000-8000-000000000041"),
		SubscriptionID: uuid.New(), Boundary: subscriptionapp.BoundaryPeriodEnd,
		RunAt:            time.Date(2026, 7, 31, 0, 0, 0, 123000000, time.UTC),
		ExpectedRevision: 7, PayloadRevision: 1,
	}
	artifact, err := subscriptionBoundaryArtifact(tenantID, intent)
	if err != nil {
		t.Fatal(err)
	}
	return tenantID, intent, subscriptionBoundaryClaimFixture{
		jobID: intent.ID, attemptID: uuid.New(), tenantID: tenantID,
		kind: subscriptionBoundaryJobKind, revision: 1,
		reference: artifact.PayloadReference,
		digest:    hex.EncodeToString(artifact.PayloadSHA256[:]), leaseGeneration: 2,
	}
}

func TestSubscriptionBoundaryArtifactRoundTripsExactClaim(t *testing.T) {
	tenantID, intent, claim := subscriptionBoundaryFixture(t)
	resolved, err := subscriptionBoundaryIntent(tenantID, claim)
	if err != nil || resolved != intent {
		t.Fatalf("boundary claim = %#v, %v", resolved, err)
	}
	second, err := subscriptionBoundaryArtifact(tenantID, intent)
	if err != nil || second.PayloadReference != claim.reference ||
		hex.EncodeToString(second.PayloadSHA256[:]) != claim.digest {
		t.Fatalf("boundary artifact drift = %#v, %v", second, err)
	}
}

func TestSubscriptionBoundaryClaimRejectsEveryAuthorityDrift(t *testing.T) {
	tenantID, _, valid := subscriptionBoundaryFixture(t)
	tests := map[string]func(*uuid.UUID, *subscriptionBoundaryClaimFixture){
		"tenant": func(candidateTenant *uuid.UUID, _ *subscriptionBoundaryClaimFixture) {
			*candidateTenant = uuid.New()
		},
		"job": func(_ *uuid.UUID, claim *subscriptionBoundaryClaimFixture) {
			claim.jobID = uuid.MustParse("019c6a7e-7c00-7000-8000-000000000042")
		},
		"kind": func(_ *uuid.UUID, claim *subscriptionBoundaryClaimFixture) {
			claim.kind = "billing.other"
		},
		"revision": func(_ *uuid.UUID, claim *subscriptionBoundaryClaimFixture) {
			claim.revision = 2
		},
		"digest": func(_ *uuid.UUID, claim *subscriptionBoundaryClaimFixture) {
			claim.digest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			claim := valid
			candidateTenant := tenantID
			mutate(&candidateTenant, &claim)
			if _, err := subscriptionBoundaryIntent(
				candidateTenant, claim,
			); !errors.Is(err, subscriptionapp.ErrRequest) {
				t.Fatalf("drift error = %v", err)
			}
		})
	}
}
