package fullrollout

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCompletionDigestBindsEveryEvidenceFieldGroup(t *testing.T) {
	mutations := map[string]func(*PersistedCompletion){
		"receipt identity":       func(v *PersistedCompletion) { v.ReceiptID[0] ^= 1 },
		"environment identity":   func(v *PersistedCompletion) { v.EnvironmentID[0] ^= 1 },
		"recorded time":          func(v *PersistedCompletion) { v.RecordedAt = v.RecordedAt.Add(time.Microsecond) },
		"step scope":             func(v *PersistedCompletion) { v.Evidence.Steps[0].Step.Scope = ScopeTenant },
		"step domain":            func(v *PersistedCompletion) { v.Evidence.Steps[0].Step.Domain += "-tampered" },
		"step scope identity":    func(v *PersistedCompletion) { v.Evidence.Steps[0].ScopeID[0] ^= 1 },
		"step requester":         func(v *PersistedCompletion) { v.Evidence.Steps[0].RequestedBy[0] ^= 1 },
		"step approver":          func(v *PersistedCompletion) { v.Evidence.Steps[0].ApprovedBy[0] ^= 1 },
		"step generation":        func(v *PersistedCompletion) { v.Evidence.Steps[0].Generation++ },
		"step state":             func(v *PersistedCompletion) { v.Evidence.Steps[0].State += "-tampered" },
		"step writer":            func(v *PersistedCompletion) { v.Evidence.Steps[0].Writer += "-tampered" },
		"step acceptance":        func(v *PersistedCompletion) { v.Evidence.Steps[0].AcceptanceSHA256[0] ^= 1 },
		"step reverse rehearsal": func(v *PersistedCompletion) { v.Evidence.Steps[0].ReverseRehearsalSHA256[0] ^= 1 },
		"step source count":      func(v *PersistedCompletion) { v.Evidence.Steps[0].SourceCount++ },
		"step target count":      func(v *PersistedCompletion) { v.Evidence.Steps[0].TargetCount++ },
		"step source digest":     func(v *PersistedCompletion) { v.Evidence.Steps[0].SourceSHA256[0] ^= 1 },
		"step target digest":     func(v *PersistedCompletion) { v.Evidence.Steps[0].TargetSHA256[0] ^= 1 },
		"step pending jobs":      func(v *PersistedCompletion) { v.Evidence.Steps[0].PendingJobs++ },
		"step pending outbox":    func(v *PersistedCompletion) { v.Evidence.Steps[0].PendingOutbox++ },
		"step pending effects":   func(v *PersistedCompletion) { v.Evidence.Steps[0].PendingEffects++ },
		"cohort identity":        func(v *PersistedCompletion) { v.Evidence.Cohorts[0].Cohort = "tampered" },
		"cohort start": func(v *PersistedCompletion) {
			v.Evidence.Cohorts[0].StartedAt = v.Evidence.Cohorts[0].StartedAt.Add(time.Microsecond)
		},
		"cohort end": func(v *PersistedCompletion) {
			v.Evidence.Cohorts[0].EndedAt = v.Evidence.Cohorts[0].EndedAt.Add(time.Microsecond)
		},
		"cohort eligible":        func(v *PersistedCompletion) { v.Evidence.Cohorts[0].EligibleRequests++ },
		"cohort compared":        func(v *PersistedCompletion) { v.Evidence.Cohorts[0].ComparedRequests++ },
		"cohort mismatch":        func(v *PersistedCompletion) { v.Evidence.Cohorts[0].MismatchCount++ },
		"cohort error metrics":   func(v *PersistedCompletion) { v.Evidence.Cohorts[0].Metrics.V2ErrorBasisPoints++ },
		"cohort latency metrics": func(v *PersistedCompletion) { v.Evidence.Cohorts[0].Metrics.V2P95Milliseconds++ },
		"cohort auth metrics":    func(v *PersistedCompletion) { v.Evidence.Cohorts[0].Metrics.AuthSuccessBasisPoints++ },
		"cohort job metrics":     func(v *PersistedCompletion) { v.Evidence.Cohorts[0].Metrics.OldestDueJobLagSeconds++ },
		"cohort DLQ metrics":     func(v *PersistedCompletion) { v.Evidence.Cohorts[0].Metrics.DLQBasisPoints++ },
		"cohort security metrics": func(v *PersistedCompletion) {
			v.Evidence.Cohorts[0].Metrics.TenantIsolationIncidents++
		},
		"evaluation time": func(v *PersistedCompletion) { v.Evidence.Now = v.Evidence.Now.Add(time.Microsecond) },
		"last legacy hit": func(v *PersistedCompletion) {
			v.Evidence.LastLegacyHitAt = v.Evidence.LastLegacyHitAt.Add(time.Microsecond)
		},
		"legacy route hits":    func(v *PersistedCompletion) { v.Evidence.LegacyRouteHits++ },
		"V1 writers":           func(v *PersistedCompletion) { v.Evidence.V1Writers++ },
		"V1 consumers":         func(v *PersistedCompletion) { v.Evidence.V1Consumers++ },
		"V1 queue messages":    func(v *PersistedCompletion) { v.Evidence.V1QueueMessages++ },
		"V1 credentials":       func(v *PersistedCompletion) { v.Evidence.V1Credentials++ },
		"shadow read artifact": func(v *PersistedCompletion) { v.Evidence.ShadowReadSHA256[0] ^= 1 },
		"logical delta artifact": func(v *PersistedCompletion) {
			v.Evidence.LogicalDeltaSHA256[0] ^= 1
		},
		"reconcile artifact":           func(v *PersistedCompletion) { v.Evidence.ReconcileSHA256[0] ^= 1 },
		"production snapshot artifact": func(v *PersistedCompletion) { v.Evidence.ProductionSnapshotSHA256[0] ^= 1 },
		"V1 removal artifact":          func(v *PersistedCompletion) { v.Evidence.V1RemovalSHA256[0] ^= 1 },
		"backup restore artifact":      func(v *PersistedCompletion) { v.Evidence.BackupRestoreSHA256[0] ^= 1 },
		"frontend contract artifact":   func(v *PersistedCompletion) { v.Evidence.FrontendContractSHA256[0] ^= 1 },
		"rollback closure artifact": func(v *PersistedCompletion) {
			v.Evidence.RollbackWindowClosureSHA256[0] ^= 1
		},
		"rollback closure": func(v *PersistedCompletion) {
			v.Evidence.RollbackWindowClosedAt = v.Evidence.RollbackWindowClosedAt.Add(time.Microsecond)
		},
	}
	baseline := validPersistedCompletion(t)
	want := digestCompletion(baseline)
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			value := validPersistedCompletion(t)
			mutate(&value)
			if digestCompletion(value) == want {
				t.Fatal("tampered evidence retained completion digest")
			}
		})
	}
}

func validPersistedCompletion(t *testing.T) PersistedCompletion {
	t.Helper()
	evidence := validCompletion(t)
	return PersistedCompletion{
		ReceiptID:     uuid.MustParse("30000000-0000-4000-8000-000000000099"),
		EnvironmentID: uuid.UUID(evidence.Steps[0].ScopeID),
		Evidence:      evidence,
		RecordedAt:    evidence.Now,
	}
}
