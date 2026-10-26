package billing

import (
	"testing"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

type providerFinalizeClaimStub struct {
	jobID, attemptID, tenantID uuid.UUID
}

func (claim providerFinalizeClaimStub) JobID() uuid.UUID     { return claim.jobID }
func (claim providerFinalizeClaimStub) AttemptID() uuid.UUID { return claim.attemptID }
func (claim providerFinalizeClaimStub) TenantID() (uuid.UUID, bool) {
	return claim.tenantID, true
}
func (providerFinalizeClaimStub) Kind() string             { return providerapp.ProviderFinalizeJobKind }
func (providerFinalizeClaimStub) PayloadRevision() int     { return 1 }
func (providerFinalizeClaimStub) PayloadReference() string { return "billing-provider-finalize/unused" }
func (providerFinalizeClaimStub) PayloadSHA256() string    { return providerTestDigest }
func (providerFinalizeClaimStub) LeaseGeneration() int64   { return 1 }

const providerTestDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestProviderFinalizationOutcomeClosesEveryDecision(t *testing.T) {
	claim := providerFinalizeClaimStub{
		jobID:     uuid.MustParse("018f1000-0000-7000-8000-000000000031"),
		attemptID: uuid.MustParse("018f1000-0000-7000-8000-000000000032"),
		tenantID:  uuid.MustParse("018f1000-0000-7000-8000-000000000033"),
	}
	tests := []struct {
		decision                    providerapp.FinalizeDecision
		state                       providerapp.CommandState
		successor, children, events int
	}{
		{providerapp.FinalizePending, providerapp.CommandQueued, 1, 0, 0},
		{providerapp.FinalizeSucceeded, providerapp.CommandSucceeded, 0, 0, 1},
		{providerapp.FinalizeFailed, providerapp.CommandFailed, 0, 0, 1},
		{providerapp.FinalizeReconcile, providerapp.CommandReconcileRequired, 1, 1, 1},
		{providerapp.FinalizeReconcileUnknown, providerapp.CommandReconcileRequired, 0, 0, 1},
		{providerapp.FinalizeReplay, providerapp.CommandSucceeded, 0, 0, 0},
	}
	for _, test := range tests {
		t.Run(string(test.decision), func(t *testing.T) {
			result := providerFinalizationFixture(test.decision, test.state)
			outcome, err := providerFinalizationOutcome(
				claim, result, providerOutcomeIDs(),
			)
			if err != nil {
				t.Fatal(err)
			}
			_, hasSuccessor := outcome.Successor()
			if boolInt(hasSuccessor) != test.successor ||
				len(outcome.Children()) != test.children ||
				len(outcome.Events()) != test.events {
				t.Fatalf("outcome successor=%v children=%d events=%d",
					hasSuccessor, len(outcome.Children()), len(outcome.Events()))
			}
			if successor, ok := outcome.Successor(); ok &&
				successor.JobID != result.FinalizerJobID {
				t.Fatalf("successor id = %s, ledger finalizer = %s",
					successor.JobID, result.FinalizerJobID)
			}
			if test.decision == providerapp.FinalizeReconcile &&
				(outcome.Children()[0].JobID == result.FinalizerJobID ||
					outcome.Children()[0].JobID != result.ActiveJobID) {
				t.Fatal("reconciliation child and successor identities are not distinct")
			}
		})
	}
}

func providerFinalizationFixture(
	decision providerapp.FinalizeDecision,
	state providerapp.CommandState,
) providerapp.Finalization {
	return providerapp.Finalization{
		Decision:       decision,
		CommandID:      uuid.MustParse("018f1000-0000-7000-8000-000000000041"),
		ActorID:        uuid.MustParse("018f1000-0000-7000-8000-000000000042"),
		ActiveJobID:    uuid.MustParse("018f1000-0000-7000-8000-000000000043"),
		FinalizerJobID: uuid.MustParse("018f1000-0000-7000-8000-000000000044"),
		Provider:       "manual", RequestSHA256: providerTestDigest, State: state,
		Revision:   2,
		OccurredAt: time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC),
	}
}

func providerOutcomeIDs() providerapp.CommandIDSource {
	values := []uuid.UUID{
		uuid.MustParse("018f1000-0000-7000-8000-000000000051"),
		uuid.MustParse("018f1000-0000-7000-8000-000000000052"),
	}
	return func() (uuid.UUID, error) {
		value := values[0]
		values = values[1:]
		return value, nil
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
