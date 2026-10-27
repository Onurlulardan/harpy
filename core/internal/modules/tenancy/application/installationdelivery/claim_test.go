package installationdelivery

import (
	"context"
	"errors"
	"testing"
)

func TestClaimRejectsInvalidCommand(t *testing.T) {
	if _, err := Claim(context.Background(), nil, ClaimCommand{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Claim() error = %v, want %v", err, ErrInvalid)
	}
}
