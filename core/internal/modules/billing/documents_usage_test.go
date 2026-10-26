package billing

import (
	"errors"
	"testing"
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func TestDocumentsUsageRequiresSealedCatalogsAndJobActor(t *testing.T) {
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	nextID := func() (uuid.UUID, error) { return uuid.NewV7() }
	clock := func() time.Time {
		return time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	}
	if _, err = NewDocumentsUsage(catalogs, nextID, clock, uuid.Nil); !errors.Is(
		err, usageapp.ErrConfiguration,
	) {
		t.Fatalf("expected invalid actor configuration, got %v", err)
	}
	actorID, _ := uuid.NewV7()
	service, err := NewDocumentsUsage(catalogs, nextID, clock, actorID)
	key, keyErr := usageapp.DocumentsStorageEntitlement()
	if err != nil || service == nil || keyErr != nil || key != "documents.storage_bytes" {
		t.Fatalf("service=%+v error=%v", service, err)
	}
}

func TestDocumentsUsageThresholdEffectFailsClosed(t *testing.T) {
	if err := (documentsUsageEffects{}).PublishThreshold(
		t.Context(), usageapp.ThresholdEventIntent{},
	); !errors.Is(err, usageapp.ErrConfiguration) {
		t.Fatalf("expected fail-closed threshold effect, got %v", err)
	}
}
