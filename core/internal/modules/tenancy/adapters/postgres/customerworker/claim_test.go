package customerworker

import (
	"context"
	"errors"
	"testing"

	workerapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerworker"
)

func TestClaimRejectsInvalidRequest(t *testing.T) {
	store := &Store{}
	if _, err := store.ClaimNext(context.Background(), workerapp.ClaimRequest{}); !errors.Is(err, workerapp.ErrInvalid) {
		t.Fatalf("ClaimNext() error = %v, want %v", err, workerapp.ErrInvalid)
	}
}
