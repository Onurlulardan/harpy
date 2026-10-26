package installationdelivery

import (
	"context"
	"errors"
	"testing"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func TestReadRejectsUnavailableStore(t *testing.T) {
	if _, err := (&commandStore{}).Read(context.Background(), installationdelivery.ReadQuery{}); !errors.Is(err, installationdelivery.ErrUnavailable) {
		t.Fatalf("Read() error = %v, want %v", err, installationdelivery.ErrUnavailable)
	}
}
