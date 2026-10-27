package postgres

import (
	"errors"
	"strings"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestJobFailureDecisionHonorsAttemptLimit(t *testing.T) {
	t.Parallel()
	_, claim := validLifecycleClaim(t, jobsv1.EffectDatabaseTransactional)
	digest := strings.Repeat("a", 64)
	retry, err := NewJobFailure(jobsv1.ErrorTransient, "DEPENDENCY_BUSY", digest, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := decideJobFailure(claim, retry)
	if err != nil || decision.state != jobsv1.StateQueued || decision.outcome != "requeued" {
		t.Fatalf("retry decision = %#v, %v", decision, err)
	}
	claim.attempt = claim.maximumAttempts
	if _, err := decideJobFailure(claim, retry); !errors.Is(err, ErrJobFailureInvalid) {
		t.Fatalf("max-attempt retry error = %v", err)
	}
	exhausted, err := NewJobFailure(jobsv1.ErrorTransient, "DEPENDENCY_BUSY", digest, 0)
	if err != nil {
		t.Fatal(err)
	}
	decision, err = decideJobFailure(claim, exhausted)
	if err != nil || decision.state != jobsv1.StateDead {
		t.Fatalf("exhausted decision = %#v, %v", decision, err)
	}
}

func TestJobFailureRejectsUnknownAndInvalidRetryEvidence(t *testing.T) {
	t.Parallel()
	digest := strings.Repeat("b", 64)
	for _, test := range []struct {
		class jobsv1.ErrorClass
		code  string
		delay time.Duration
	}{
		{jobsv1.ErrorPermanent, "PERMANENT_FAILURE", time.Second},
		{jobsv1.ErrorTransient, jobsv1.UnknownOutcomeCode, 0},
		{jobsv1.ErrorConflict, "CONFLICT", 15*time.Minute + time.Nanosecond},
	} {
		if _, err := NewJobFailure(test.class, test.code, digest, test.delay); !errors.Is(
			err, ErrJobFailureInvalid,
		) {
			t.Fatalf("invalid failure %#v error = %v", test, err)
		}
	}
}
