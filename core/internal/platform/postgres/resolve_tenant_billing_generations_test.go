package postgres

import (
	"strings"
	"testing"
)

func TestResolveTenantBillingGenerationsSQLRequiresExactCurrentGroup(t *testing.T) {
	for _, fragment := range []string{
		"group_row.scope_kind = 'tenant'",
		"group_row.group_id = 'billing'",
		"billing_member.domain = 'billing'",
		"group_row.state = 'v2'",
		"billing_row.writer = 'v2'",
		"extra.domain <> 'billing'",
	} {
		if !strings.Contains(resolveTenantBillingGenerationsSQL, fragment) {
			t.Fatalf("tenant Billing generation SQL missing %q", fragment)
		}
	}
}
