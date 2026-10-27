package installationdelivery

import (
	"context"
	"errors"
	"testing"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func TestDeleteRejectsUnavailableStore(t *testing.T) {
	if _, err := (&commandStore{}).DeleteSecret(context.Background(), installationdelivery.StateCommand{}); !errors.Is(err, installationdelivery.ErrUnavailable) {
		t.Fatalf("DeleteSecret() error = %v, want %v", err, installationdelivery.ErrUnavailable)
	}
}
