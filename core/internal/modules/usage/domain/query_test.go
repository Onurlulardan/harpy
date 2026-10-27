package domain

import (
	"strings"
	"testing"
	"time"
)

func TestTimeRangeAndRankedQueryAreBounded(t *testing.T) {
	end := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	window, err := NewTimeRange(end.Add(-24*time.Hour), end)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = NewRankedQuery(window, 50, 50); err != nil {
		t.Fatal(err)
	}
	if _, err = NewTimeRange(end.Add(-MaximumWindow-time.Microsecond), end); err == nil {
		t.Fatal("unbounded time range was accepted")
	}
	if _, err = NewRankedQuery(window, 51, 50); err == nil {
		t.Fatal("rank limit above maximum was accepted")
	}
}

func TestTimelineSubjectMatchesScope(t *testing.T) {
	tenantID := "10000000-0000-4000-8000-000000000001"
	if _, err := NewTimelineQuery(TenantScope, tenantID, 1, 100); err != nil {
		t.Fatal(err)
	}
	if _, err := NewTimelineQuery(PlatformScope, strings.Repeat("a", 64), 1, 20); err != nil {
		t.Fatal(err)
	}
	if _, err := NewTimelineQuery(PlatformScope, tenantID, 1, 20); err == nil {
		t.Fatal("raw UUID was accepted by platform lookup")
	}
}
