package postgres

import (
	"strings"
	"testing"
)

func TestAuthorizeSQLUsesOnlyTenantEvidenceRoutine(t *testing.T) {
	required := []string{
		"iam_authorization_evidence_read_v1",
		"(pg_catalog.current_schemas(false))[2]::name",
		"$1::uuid", "$2::text", "$3::text",
	}
	for _, marker := range required {
		if !strings.Contains(authorizeSQL, marker) {
			t.Fatalf("authorize SQL omits %s", marker)
		}
	}
	for _, forbidden := range []string{
		"iam_users ", "iam_permissions ", "iam_authorization_clock ",
		"platform.", "public.", "pg_temp", "current_schema()", "$4",
	} {
		if strings.Contains(authorizeSQL, forbidden) {
			t.Fatalf("authorize SQL contains forbidden token %q", forbidden)
		}
	}
}
