package alertrule

import (
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestUpdateForwardsAlertRule(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewAlertRuleOperations(backend).UpdateAlertRule(
		t.Context(), "rule", platformopsapi.AlertRuleInput{})
	if err != nil || backend.operation != "update" || result.ID != "rule" {
		t.Fatalf("UpdateAlertRule() = %#v, %v", result, err)
	}
}
