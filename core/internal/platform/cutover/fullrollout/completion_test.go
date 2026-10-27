package fullrollout

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
)

func TestCompletionRequiresExactDomainOrderAndAllV1SurfacesAtZero(t *testing.T) {
	evidence := validCompletion(t)
	if err := EvaluateCompletion(evidence); err != nil {
		t.Fatal(err)
	}
	evidence.Steps[4], evidence.Steps[5] = evidence.Steps[5], evidence.Steps[4]
	if err := EvaluateCompletion(evidence); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("reordered domains error = %v", err)
	}
	evidence = validCompletion(t)
	evidence.V1Consumers = 1
	if err := EvaluateCompletion(evidence); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("remaining V1 consumer error = %v", err)
	}
}

func TestCompletionRejectsShortSilenceAndCrossTenantEvidence(t *testing.T) {
	evidence := validCompletion(t)
	evidence.LastLegacyHitAt = evidence.Now.Add(-LegacySilenceWindow + time.Second)
	if err := EvaluateCompletion(evidence); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("short silence error = %v", err)
	}
	evidence = validCompletion(t)
	evidence.Steps[8].ScopeID[0] ^= 0xff
	if err := EvaluateCompletion(evidence); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("cross-tenant evidence error = %v", err)
	}
}

func validCompletion(t *testing.T) CompletionEvidence {
	t.Helper()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	digest := sha256.Sum256([]byte("verified evidence"))
	principal := [16]byte{1}
	approver := [16]byte{2}
	environment := [16]byte{3}
	tenant := [16]byte{4}
	steps := make([]StepEvidence, len(requiredSteps))
	for index, step := range requiredSteps {
		scopeID := tenant
		if step.Scope == ScopeGlobal {
			scopeID = environment
		}
		steps[index] = StepEvidence{Step: step, ScopeID: scopeID,
			RequestedBy: principal, ApprovedBy: approver, Generation: int64(index + 1),
			State: "v2", Writer: "v2", AcceptanceSHA256: digest,
			ReverseRehearsalSHA256: digest, SourceCount: 10, TargetCount: 10,
			SourceSHA256: digest, TargetSHA256: digest}
	}
	order := grouprehearsal.CohortOrder()
	var total time.Duration
	for _, cohort := range order {
		minimum, _ := grouprehearsal.MinimumSoak(cohort)
		total += minimum
	}
	cursor := now.Add(-total - time.Hour)
	cohorts := make([]grouprehearsal.SoakEvidence, 0, len(order))
	for _, cohort := range order {
		minimum, _ := grouprehearsal.MinimumSoak(cohort)
		end := cursor.Add(minimum)
		cohorts = append(cohorts, grouprehearsal.SoakEvidence{Cohort: cohort,
			StartedAt: cursor, EndedAt: end, EligibleRequests: 10_000,
			ComparedRequests: 10_000, Metrics: grouprehearsal.Metrics{}})
		cursor = end
	}
	return CompletionEvidence{Steps: steps, Cohorts: cohorts, Now: now,
		LastLegacyHitAt: now.Add(-LegacySilenceWindow), ShadowReadSHA256: digest,
		LogicalDeltaSHA256: digest, ReconcileSHA256: digest,
		ProductionSnapshotSHA256: digest, V1RemovalSHA256: digest,
		BackupRestoreSHA256:         digest,
		FrontendContractSHA256:      digest,
		RollbackWindowClosureSHA256: digest,
		RollbackWindowClosedAt:      now.Add(-30 * time.Minute)}
}
