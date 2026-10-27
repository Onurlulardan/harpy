package postgres

import (
	"strings"
	"testing"
)

func TestResolveTenantNotificationGenerationsSQLRequiresExactCurrentGroup(t *testing.T) {
	fragments := []string{
		"group_row.scope_kind = 'tenant'",
		"group_row.group_id = 'notifications'",
		"notification_member.domain = 'notifications'",
		"notification_member.ordinal = 1",
		"group_row.state = 'v2'",
		"group_row.writer = 'v2'",
		"notification_row.state = 'v2'",
		"notification_row.writer = 'v2'",
		"extra.domain <> 'notifications'",
	}
	for _, fragment := range fragments {
		if !strings.Contains(resolveTenantNotificationGenerationsSQL, fragment) {
			t.Fatalf("tenant Notifications generation SQL missing %q", fragment)
		}
	}
}
