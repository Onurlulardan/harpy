package async

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestDecideLeaseRecoveryFreezesEveryEffectBoundary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		cancelRequested  bool
		effectStarted    bool
		hasEffectReceipt bool
		wantState        jobsv1.State
		wantCode         string
		wantReuseReceipt bool
	}{
		{"safe retry", false, false, false, jobsv1.StateQueued, "", false},
		{"cancel before effect", true, false, false, jobsv1.StateCancelled, "", false},
		{"known effect", false, true, true, jobsv1.StateQueued, "", true},
		{"known effect ignores cancel", true, true, true, jobsv1.StateQueued, "", true},
		{"unknown outcome", false, true, false, jobsv1.StateDead, jobsv1.UnknownOutcomeCode, false},
		{"unknown outcome ignores cancel", true, true, false, jobsv1.StateDead, jobsv1.UnknownOutcomeCode, false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, err := DecideLeaseRecovery(
				test.cancelRequested, test.effectStarted, test.hasEffectReceipt,
			)
			if err != nil {
				t.Fatal(err)
			}
			if actual.State != test.wantState || actual.Code != test.wantCode ||
				actual.ReuseReceipt != test.wantReuseReceipt {
				t.Fatalf("decision = %#v", actual)
			}
		})
	}
}

func TestDecideLeaseRecoveryRejectsReceiptBeforeEffect(t *testing.T) {
	t.Parallel()
	if _, err := DecideLeaseRecovery(false, false, true); err == nil {
		t.Fatal("receipt without effect start was accepted")
	}
}
