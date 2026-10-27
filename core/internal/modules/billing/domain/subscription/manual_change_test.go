package subscription

import (
	"testing"
	"time"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
)

func TestManualChangePinsTargetAtPeriodEndWithoutProration(t *testing.T) {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	period, _ := billingperiod.New(start, end)
	change, err := ScheduleManualChange(
		Active, period, start, "USD", "USD", versionID(1), versionID(2))
	if err != nil || !change.EffectiveAt().Equal(end) || change.Currency() != "USD" ||
		change.OfferingVersionID() != versionID(1) || change.PriceVersionID() != versionID(2) {
		t.Fatalf("change=%+v err=%v", change, err)
	}
}

func TestManualChangeRejectsCurrencyConversionAndStaleBoundary(t *testing.T) {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	period, _ := billingperiod.New(start, end)
	for _, input := range []struct {
		now             time.Time
		current, target string
	}{
		{start, "USD", "EUR"},
		{end, "USD", "USD"},
		{start, "usd", "usd"},
	} {
		if _, err := ScheduleManualChange(
			Active, period, input.now, input.current, input.target,
			versionID(1), versionID(2),
		); err != ErrChange {
			t.Fatalf("input %+v: expected ErrChange, got %v", input, err)
		}
	}
}

func versionID(last byte) VersionIdentifier {
	var identifier VersionIdentifier
	identifier[len(identifier)-1] = last
	return identifier
}
