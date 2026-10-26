package async

import (
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestDecideFailureAppliesRetryAndEffectBoundaries(t *testing.T) {
	t.Parallel()
	external := failurePolicy(t, jobsv1.EffectExternal, 3)
	database := failurePolicy(t, jobsv1.EffectDatabaseTransactional, 3)
	zero := fixedJitter{choose: func(uint64) uint64 { return 0 }}
	tests := []struct {
		name      string
		policy    jobsv1.KindPolicy
		attempt   int
		class     jobsv1.ErrorClass
		code      string
		started   bool
		receipt   bool
		wantState jobsv1.State
		wantCode  string
		wantError bool
	}{
		{"transient retry", database, 1, jobsv1.ErrorTransient, "JOB_TRANSIENT", false, false, jobsv1.StateQueued, "JOB_TRANSIENT", false},
		{"retry exhausted", database, 3, jobsv1.ErrorConflict, "JOB_CONFLICT", false, false, jobsv1.StateDead, "JOB_CONFLICT", false},
		{"permanent", external, 1, jobsv1.ErrorPermanent, "JOB_REJECTED", true, true, jobsv1.StateDead, "JOB_REJECTED", false},
		{"cancelled", database, 1, jobsv1.ErrorCancelled, "JOB_CANCELLED", false, false, jobsv1.StateCancelled, "JOB_CANCELLED", false},
		{"unknown outcome", external, 1, jobsv1.ErrorTransient, "JOB_TRANSIENT", true, false, jobsv1.StateDead, jobsv1.UnknownOutcomeCode, false},
		{"receipt before effect", external, 1, jobsv1.ErrorPermanent, "JOB_REJECTED", false, true, "", "", true},
		{"transactional effect", database, 1, jobsv1.ErrorPermanent, "JOB_REJECTED", true, true, "", "", true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, err := DecideFailure(
				test.policy, test.attempt, test.class, test.code,
				test.started, test.receipt, zero,
			)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v", err)
			}
			if err == nil && (actual.State != test.wantState ||
				actual.Code != test.wantCode || actual.Delay != 0) {
				t.Fatalf("decision = %#v", actual)
			}
		})
	}
}

func TestDecideFailureReturnsChosenFullJitter(t *testing.T) {
	t.Parallel()
	policy := failurePolicy(t, jobsv1.EffectDatabaseTransactional, 2)
	decision, err := DecideFailure(
		policy, 1, jobsv1.ErrorTransient, "JOB_TRANSIENT", false, false,
		fixedJitter{choose: func(uint64) uint64 { return uint64(2 * time.Second) }},
	)
	if err != nil {
		t.Fatal(err)
	}
	if decision.State != jobsv1.StateQueued || decision.Delay != 2*time.Second {
		t.Fatalf("decision = %#v", decision)
	}
}

func failurePolicy(t *testing.T, effect jobsv1.EffectClass, attempts int) jobsv1.KindPolicy {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy("test.job", 1, "forms", effect, attempts)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}
