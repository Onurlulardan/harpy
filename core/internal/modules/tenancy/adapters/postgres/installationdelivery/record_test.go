package installationdelivery

import (
	"context"
	"errors"
	"testing"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func TestRecordRejectsUnavailableStore(t *testing.T) {
	if _, err := (&commandStore{}).RecordUnknownOutcome(context.Background(), installationdelivery.UnknownCommand{}); !errors.Is(err, installationdelivery.ErrUnavailable) {
		t.Fatalf("RecordUnknownOutcome() error = %v, want %v", err, installationdelivery.ErrUnavailable)
	}
}
