package dashboard

import (
	"context"
	"errors"
	"strings"
	"testing"

	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
)

func TestReadDefaultFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).ReadDefault(context.Background(), "home")
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}

func TestReadDefaultSQLUsesOwnerRoutineOnly(t *testing.T) {
	if !strings.Contains(getDefaultSQL, "analytics_dashboard_get_default_v1") {
		t.Fatal("default dashboard read does not use its owner routine")
	}
	if strings.Contains(getDefaultSQL, "analytics_dashboard_preferences") {
		t.Fatal("default dashboard read accesses its owner table directly")
	}
}
