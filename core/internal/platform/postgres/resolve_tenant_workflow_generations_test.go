package postgres

import (
	"strings"
	"testing"
)

func TestResolveTenantWorkflowGenerationsSQLRequiresExactCurrentGroup(t *testing.T) {
	for _, fragment := range []string{
		"group_row.scope_kind = 'tenant'",
		"group_row.group_id = 'workflow'",
		"workflow_member.domain = 'workflow'",
		"workflow_member.ordinal = 1",
		"group_row.state = 'v2'",
		"group_row.writer = 'v2'",
		"workflow_row.state = 'v2'",
		"workflow_row.writer = 'v2'",
		"extra.domain <> 'workflow'",
	} {
		if !strings.Contains(resolveTenantWorkflowGenerationsSQL, fragment) {
			t.Fatalf("tenant Workflow generation SQL missing %q", fragment)
		}
	}
}
