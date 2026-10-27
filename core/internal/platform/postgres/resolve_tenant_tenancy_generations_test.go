package postgres

import (
	"strings"
	"testing"
)

func TestResolveTenantTenancyGenerationsSQLRequiresExactCurrentGroup(t *testing.T) {
	fragments := []string{
		"group_row.scope_kind = 'tenant'",
		"group_row.group_id = 'tenancy'",
		"tenancy_member.domain = 'tenancy'",
		"tenancy_member.ordinal = 1",
		"group_row.state = 'v2'",
		"group_row.writer = 'v2'",
		"tenancy_row.state = 'v2'",
		"tenancy_row.writer = 'v2'",
		"extra.domain <> 'tenancy'",
	}
	for _, fragment := range fragments {
		if !strings.Contains(resolveTenantTenancyGenerationsSQL, fragment) {
			t.Fatalf("tenant tenancy generation SQL missing %q", fragment)
		}
	}
}
