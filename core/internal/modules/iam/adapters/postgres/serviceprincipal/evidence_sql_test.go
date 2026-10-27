package serviceprincipal

import (
	"strings"
	"testing"
)

func TestEvidenceSQLUsesOnlyTenantReadRoutine(t *testing.T) {
	for _, marker := range []string{
		"iam_capability_grant_read_v1",
		"(pg_catalog.current_schemas(false))[2]::name",
		"$1::uuid", "$2::bigint",
	} {
		if !strings.Contains(evidenceSQL, marker) {
			t.Fatalf("evidence SQL lacks %q", marker)
		}
	}
	for _, direct := range []string{
		"iam_service_principals AS", "iam_capability_grants AS",
		"iam_capability_grant_revisions AS", "current_schema()",
	} {
		if strings.Contains(evidenceSQL, direct) {
			t.Fatalf("evidence SQL directly reads %q", direct)
		}
	}
}
