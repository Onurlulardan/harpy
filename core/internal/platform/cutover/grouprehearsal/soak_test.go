package grouprehearsal

import (
	"testing"
	"time"
)

func TestEveryCohortEnforcesItsMinimumSoak(t *testing.T) {
	start := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	for _, cohort := range CohortOrder() {
		minimum, _ := MinimumSoak(cohort)
		value := validSoak(cohort, start, minimum)
		if err := EvaluateSoak(value); err != nil {
			t.Fatalf("%s: %v", cohort, err)
		}
		value.EndedAt = value.EndedAt.Add(-time.Nanosecond)
		if err := EvaluateSoak(value); err == nil {
			t.Fatalf("%s short soak accepted", cohort)
		}
	}
}

func TestEverySafetyThresholdFencesTheGroup(t *testing.T) {
	start := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	mutations := []func(*Metrics){
		func(value *Metrics) { value.TenantIsolationIncidents = 1 },
		func(value *Metrics) { value.V2ErrorBasisPoints = 51 },
		func(value *Metrics) { value.BaselineErrorBasisPoints, value.V2ErrorBasisPoints = 20, 41 },
		func(value *Metrics) { value.V2P95Milliseconds = 121 },
		func(value *Metrics) { value.AuthSuccessBasisPoints = 9979 },
		func(value *Metrics) { value.OldestDueJobLagSeconds = 61 },
		func(value *Metrics) { value.DLQBasisPoints = 11 },
		func(value *Metrics) { value.PermanentSecurityJobs = 1 },
	}
	for index, mutate := range mutations {
		value := validSoak(CohortInternal, start, 24*time.Hour)
		mutate(&value.Metrics)
		if err := EvaluateSoak(value); err != ErrFence {
			t.Fatalf("threshold %d error = %v", index, err)
		}
	}
}

func validSoak(cohort Cohort, start time.Time, duration time.Duration) SoakEvidence {
	return SoakEvidence{Cohort: cohort, StartedAt: start, EndedAt: start.Add(duration),
		EligibleRequests: 10_000, ComparedRequests: 10_000,
		Metrics: Metrics{BaselineErrorBasisPoints: 20, V2ErrorBasisPoints: 20,
			BaselineP95Milliseconds: 100, V2P95Milliseconds: 100,
			BaselineAuthSuccessBasisPoints: 10_000, AuthSuccessBasisPoints: 10_000,
			BaselineJobLagSeconds: 10, OldestDueJobLagSeconds: 10}}
}
