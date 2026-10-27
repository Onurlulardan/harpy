package alertrule

import (
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestCreateForwardsAlertRule(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewAlertRuleOperations(backend).CreateAlertRule(
		t.Context(), platformopsapi.AlertRuleInput{})
	if err != nil || backend.operation != "create" || result.ID != "rule" {
		t.Fatalf("CreateAlertRule() = %#v, %v", result, err)
	}
}
