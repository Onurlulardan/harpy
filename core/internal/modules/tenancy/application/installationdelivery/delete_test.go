package installationdelivery

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteRejectsInvalidCommand(t *testing.T) {
	if _, err := DeleteSecret(context.Background(), nil, StateCommand{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("DeleteSecret() error = %v, want %v", err, ErrInvalid)
	}
}
