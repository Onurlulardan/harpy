package installationdelivery

import (
	"context"
	"errors"
	"testing"
)

func TestRecordRejectsInvalidCommand(t *testing.T) {
	if _, err := RecordUnknownOutcome(context.Background(), nil, UnknownCommand{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("RecordUnknownOutcome() error = %v, want %v", err, ErrInvalid)
	}
}
