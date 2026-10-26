package postgres

import (
	"strings"
	"testing"
)

func TestAssigneeAccessLocksAndRequiresEveryPrincipal(t *testing.T) {
	for _, fragment := range []string{"FOR SHARE", "NOT EXISTS", "membership.active",
		"membership.principal_id IS NULL"} {
		if !strings.Contains(assigneesHaveAccessSQL, fragment) {
			t.Fatalf("assignee access query missing %q", fragment)
		}
	}
}
