package postgres

import (
	"strings"
	"testing"
)

func TestResolveTenantWorkGenerationsRequiresExactAuthority(t *testing.T) {
	for _, fragment := range []string{
		"group_row.group_id = 'work'", "work_member.domain = 'work'",
		"work_member.ordinal = 1", "group_row.state = 'v2'",
		"group_row.writer = 'v2'", "work_row.state = 'v2'",
		"work_row.writer = 'v2'", "extra.domain <> 'work'",
	} {
		if !strings.Contains(resolveTenantWorkGenerationsSQL, fragment) {
			t.Fatalf("Work generation SQL misses %q", fragment)
		}
	}
}
