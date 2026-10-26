package postgres

import (
	"strings"
	"testing"
)

func TestGlobalBillingGenerationProjectionIsClosedAndWriterFenced(t *testing.T) {
	required := []string{
		"platform.environment_identity", "group_row.group_id = 'billing'",
		"member_row.domain = 'billing'", "group_row.state = 'v2'",
		"group_row.writer = 'v2'", "domain_row.state = 'v2'",
		"domain_row.writer = 'v2'", "NOT EXISTS",
	}
	for _, fragment := range required {
		if !strings.Contains(resolveGlobalBillingGenerationsSQL, fragment) {
			t.Fatalf("global Billing projection lacks %q", fragment)
		}
	}
}
