package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/fullrollout"
)

func TestRecordFullRolloutRejectsMissingProtectedRecorder(t *testing.T) {
	if _, err := RecordFullRollout(
		context.Background(), nil, fullrollout.PersistedCompletion{},
	); !errors.Is(err, fullrollout.ErrInvalidEvidence) {
		t.Fatalf("missing protected recorder error = %v", err)
	}
}
