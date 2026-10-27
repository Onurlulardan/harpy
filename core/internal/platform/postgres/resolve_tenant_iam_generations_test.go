package postgres

import (
	"strings"
	"testing"
)

func TestResolveTenantIAMGenerationsSQLRequiresExactCurrentGroup(t *testing.T) {
	fragments := []string{
		"group_row.group_id = 'iam_system_resources'",
		"iam_member.domain = 'iam'",
		"iam_member.ordinal = 1",
		"layout_member.domain = 'system_resource_layout'",
		"layout_member.ordinal = 2",
		"group_row.state = 'v2' AND group_row.writer = 'v2'",
		"extra.domain NOT IN ('iam', 'system_resource_layout')",
	}
	for _, fragment := range fragments {
		if !strings.Contains(resolveTenantIAMGenerationsSQL, fragment) {
			t.Fatalf("IAM generation SQL missing %q", fragment)
		}
	}
}
