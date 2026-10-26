package postgres

import (
	"strings"
	"testing"
)

func TestRecordCompatibilityHitIsDurableAndRetirementAware(t *testing.T) {
	for _, fragment := range []string{
		"ON CONFLICT (route_contract, consumer_id, build_id)",
		"hit_count = work_compatibility_hits.hit_count + 1",
		"retired_at IS NULL",
	} {
		if !strings.Contains(recordCompatibilityHitSQL, fragment) {
			t.Fatalf("compatibility hit query missing %q", fragment)
		}
	}
}
