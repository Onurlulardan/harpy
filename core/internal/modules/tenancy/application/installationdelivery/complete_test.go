package installationdelivery

import (
	"context"
	"errors"
	"testing"
)

func TestCompleteRejectsInvalidCommand(t *testing.T) {
	if _, err := Complete(context.Background(), nil, StateCommand{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Complete() error = %v, want %v", err, ErrInvalid)
	}
}
