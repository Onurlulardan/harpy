package fullrollout

import (
	"errors"
	"testing"
	"time"
)

func TestCompletionRequiresEveryFinalArtifactAndZeroLegacySurface(t *testing.T) {
	mutations := map[string]func(*CompletionEvidence){
		"shadow read":         func(v *CompletionEvidence) { v.ShadowReadSHA256 = [32]byte{} },
		"logical delta":       func(v *CompletionEvidence) { v.LogicalDeltaSHA256 = [32]byte{} },
		"reconcile":           func(v *CompletionEvidence) { v.ReconcileSHA256 = [32]byte{} },
		"production snapshot": func(v *CompletionEvidence) { v.ProductionSnapshotSHA256 = [32]byte{} },
		"V1 removal":          func(v *CompletionEvidence) { v.V1RemovalSHA256 = [32]byte{} },
		"backup restore":      func(v *CompletionEvidence) { v.BackupRestoreSHA256 = [32]byte{} },
		"frontend contract":   func(v *CompletionEvidence) { v.FrontendContractSHA256 = [32]byte{} },
		"MCP contract":        func(v *CompletionEvidence) { v.MCPContractSHA256 = [32]byte{} },
		"rollback closure":    func(v *CompletionEvidence) { v.RollbackWindowClosureSHA256 = [32]byte{} },
		"legacy route hit":    func(v *CompletionEvidence) { v.LegacyRouteHits = 1 },
		"V1 writer":           func(v *CompletionEvidence) { v.V1Writers = 1 },
		"V1 consumer":         func(v *CompletionEvidence) { v.V1Consumers = 1 },
		"V1 queue message":    func(v *CompletionEvidence) { v.V1QueueMessages = 1 },
		"V1 credential":       func(v *CompletionEvidence) { v.V1Credentials = 1 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			value := validCompletion(t)
			mutate(&value)
			if err := EvaluateCompletion(value); !errors.Is(err, ErrInvalidEvidence) {
				t.Fatalf("missing or non-zero evidence error = %v", err)
			}
		})
	}
}

func TestCompletionRequiresChronologicalCohortsBeforeRollbackClosure(t *testing.T) {
	value := validCompletion(t)
	value.Cohorts[1].StartedAt = value.Cohorts[0].EndedAt.Add(-time.Microsecond)
	if err := EvaluateCompletion(value); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("overlapping cohort error = %v", err)
	}

	value = validCompletion(t)
	last := value.Cohorts[len(value.Cohorts)-1]
	value.RollbackWindowClosedAt = last.EndedAt.Add(-time.Microsecond)
	if err := EvaluateCompletion(value); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("early rollback closure error = %v", err)
	}

	value = validCompletion(t)
	value.RollbackWindowClosedAt = value.LastLegacyHitAt.Add(-time.Microsecond)
	if err := EvaluateCompletion(value); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("rollback closure before legacy silence error = %v", err)
	}
}
