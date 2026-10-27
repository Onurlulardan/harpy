package alertevent

import (
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestListForwardsAlertEventFilter(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewAlertEventOperations(backend).ListAlertEvents(
		t.Context(), "environment", platformopsapi.AlertEventFilter{})
	if err != nil || backend.operation != "list" || result.Total != 1 {
		t.Fatalf("ListAlertEvents() = %#v, %v", result, err)
	}
}
