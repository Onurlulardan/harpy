package installationdelivery

import (
	"context"
	"errors"
	"testing"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func TestCompleteRejectsUnavailableStore(t *testing.T) {
	if _, err := (&commandStore{}).Complete(context.Background(), installationdelivery.StateCommand{}); !errors.Is(err, installationdelivery.ErrUnavailable) {
		t.Fatalf("Complete() error = %v, want %v", err, installationdelivery.ErrUnavailable)
	}
}
