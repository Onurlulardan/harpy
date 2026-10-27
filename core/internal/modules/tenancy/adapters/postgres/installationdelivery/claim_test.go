package installationdelivery

import (
	"context"
	"errors"
	"testing"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func TestClaimRejectsUnavailableStore(t *testing.T) {
	if _, err := (&commandStore{}).Claim(context.Background(), installationdelivery.ClaimCommand{}); !errors.Is(err, installationdelivery.ErrUnavailable) {
		t.Fatalf("Claim() error = %v, want %v", err, installationdelivery.ErrUnavailable)
	}
}
