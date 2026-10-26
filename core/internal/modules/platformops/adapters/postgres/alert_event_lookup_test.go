package postgres

import (
	"testing"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestAlertEventLookupScansOwnedEvent(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	database := &operationDatabase{row: func(destinations ...any) error {
		*destinations[0].(*string) = "event"
		*destinations[1].(*string) = "environment"
		rule := "rule"
		*destinations[2].(**string) = &rule
		*destinations[3].(*string) = "core-api"
		*destinations[4].(*platformopsapi.HealthStatus) = platformopsapi.HealthHealthy
		*destinations[5].(*platformopsapi.HealthStatus) = platformopsapi.HealthUnhealthy
		*destinations[6].(*[]string) = []string{}
		*destinations[7].(*string) = "probe"
		*destinations[8].(**time.Time) = nil
		*destinations[9].(*time.Time) = now
		return nil
	}}
	result, err := operationStore(database).AlertEventLookup(
		t.Context(), "environment", "event")
	if err != nil || result.ID != "event" || result.RuleID != "rule" {
		t.Fatalf("AlertEventLookup() = %#v, %v", result, err)
	}
}
