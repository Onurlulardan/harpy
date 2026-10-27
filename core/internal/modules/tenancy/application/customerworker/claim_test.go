package customerworker

import (
	"context"
	"errors"
	"testing"
)

func TestClaimRejectsInvalidRequest(t *testing.T) {
	if _, err := ClaimNext(context.Background(), claimStoreFixture{}, ClaimRequest{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ClaimNext() error = %v, want %v", err, ErrInvalid)
	}
}

type claimStoreFixture struct{}

func (claimStoreFixture) ClaimNext(context.Context, ClaimRequest) (Claim, error) {
	return Claim{}, ErrNoWork
}
