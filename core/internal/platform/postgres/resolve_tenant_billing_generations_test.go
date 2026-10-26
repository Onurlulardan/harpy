package postgres

import (
	"strings"
	"testing"
)

func TestResolveTenantBillingGenerationsSQLRequiresExactCurrentGroup(t *testing.T) {
	for _, fragment := range []string{
		"group_row.scope_kind = 'tenant'",
		"group_row.group_id = 'billing_forms'",
		"billing_member.domain = 'billing'",
		"forms_member.domain = 'forms'",
		"group_row.state = 'v2'",
		"billing_row.writer = 'v2'",
		"forms_row.writer = 'v2'",
		"extra.domain NOT IN ('billing', 'forms')",
	} {
		if !strings.Contains(resolveTenantBillingGenerationsSQL, fragment) {
			t.Fatalf("tenant Billing generation SQL missing %q", fragment)
		}
	}
}
