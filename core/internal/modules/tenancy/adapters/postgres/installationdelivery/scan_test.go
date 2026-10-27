package installationdelivery

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

type deliveryRow func(...any) error

func (row deliveryRow) Scan(destinations ...any) error { return row(destinations...) }

func TestScanDeliveryNormalizesDatabaseTimeToUTC(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.FixedZone("local", 3*60*60))
	row := deliveryRow(func(destinations ...any) error {
		*destinations[0].(*uuid.UUID) = uuid.New()
		*destinations[1].(*uuid.UUID) = uuid.New()
		*destinations[2].(*uuid.UUID) = uuid.New()
		*destinations[3].(*uuid.UUID) = uuid.New()
		*destinations[4].(*uuid.UUID) = uuid.New()
		*destinations[5].(*int64) = 1
		*destinations[6].(*string) = "ready_to_display"
		*destinations[7].(*int64) = 1
		*destinations[8].(*time.Time) = now
		return nil
	})
	delivery, err := scanDelivery(row)
	if err != nil {
		t.Fatalf("scanDelivery() error = %v", err)
	}
	if delivery.DatabaseNow.Location() != time.UTC || !delivery.Valid() {
		t.Fatalf("scanDelivery() database time = %v, delivery valid = %t",
			delivery.DatabaseNow.Location(), delivery.Valid())
	}
}

func TestScanTransitionNormalizesDatabaseTimeToUTC(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.FixedZone("local", 3*60*60))
	row := deliveryRow(func(destinations ...any) error {
		*destinations[0].(*string) = "completed"
		*destinations[1].(*int64) = 5
		*destinations[2].(*time.Time) = now
		return nil
	})
	transition, err := scanTransition(row)
	if err != nil {
		t.Fatalf("scanTransition() error = %v", err)
	}
	if transition.DatabaseNow.Location() != time.UTC ||
		transition.State != "completed" || transition.Revision != 5 {
		t.Fatalf("scanTransition() = %+v", transition)
	}
}
