package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func TestChangeOperationRejectsMissingStore(t *testing.T) {
	_, err := ChangeOperation(context.Background(), nil, OperationMutation{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}

func TestDurableResumeFailureMarksInterruptedResumeFailed(t *testing.T) {
	for _, cause := range []error{
		ErrResumeInterrupted, context.Canceled, context.DeadlineExceeded,
	} {
		state, code, durable := durableResumeFailure(cause)
		if !durable || state != catalogdomain.OperationFailed ||
			code != "CATALOG_RESUME_INTERRUPTED" {
			t.Fatalf("cause=%v state=%s code=%s durable=%v", cause, state, code, durable)
		}
	}
	if _, _, durable := durableResumeFailure(errors.New("transient")); durable {
		t.Fatal("unknown resume error was treated as durable")
	}
	if ResumeFailureDurable(errors.New("transient")) {
		t.Fatal("unknown resume error was durable")
	}
	if !ResumeFailureDurable(ErrResumeInterrupted) {
		t.Fatal("interrupted resume was not durable")
	}
}
