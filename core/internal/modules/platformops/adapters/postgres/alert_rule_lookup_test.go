package postgres

import (
	"testing"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestAlertRuleLookupScansOwnedRule(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	database := &operationDatabase{row: func(destinations ...any) error {
		*destinations[0].(*string) = "rule"
		*destinations[1].(*string) = "environment"
		*destinations[2].(*string) = "availability"
		*destinations[3].(*string) = "description"
		*destinations[4].(*bool) = true
		service := "core-api"
		*destinations[5].(**string) = &service
		*destinations[6].(*platformopsapi.AlertCondition) = platformopsapi.AlertUnhealthy
		*destinations[7].(*int) = 5
		*destinations[8].(*[]byte) = []byte(`[]`)
		*destinations[9].(*time.Time) = now
		*destinations[10].(*time.Time) = now
		return nil
	}}
	result, err := operationStore(database).AlertRuleLookup(
		t.Context(), "environment", "rule")
	if err != nil || result.ID != "rule" || result.EnvironmentID != "environment" {
		t.Fatalf("AlertRuleLookup() = %#v, %v", result, err)
	}
}
