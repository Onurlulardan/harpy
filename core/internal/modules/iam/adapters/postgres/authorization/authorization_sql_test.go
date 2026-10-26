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

func TestFormsAuthorizeSQLUsesFencedV2EvidenceRoutine(t *testing.T) {
	required := []string{
		"iam_authorization_evidence_read_v2",
		"(pg_catalog.current_schemas(false))[2]::name",
		"$1::uuid", "$2::uuid", "$3::text", "$4::text",
	}
	for _, marker := range required {
		if !strings.Contains(authorizeFormsSQL, marker) {
			t.Fatalf("Forms authorize SQL omits %s", marker)
		}
	}
	for _, forbidden := range []string{
		"iam_users ", "iam_permissions ", "iam_authorization_clock ",
		"platform.", "public.", "pg_temp", "current_schema()", "$5",
	} {
		if strings.Contains(authorizeFormsSQL, forbidden) {
			t.Fatalf("Forms authorize SQL contains forbidden token %q", forbidden)
		}
	}
}

func TestFormsSnapshotAuthorizeSQLUsesReadOnlyV3EvidenceRoutine(t *testing.T) {
	required := []string{
		"iam_authorization_evidence_read_v3",
		"(pg_catalog.current_schemas(false))[2]::name",
		"$1::uuid", "$2::uuid", "$3::text", "$4::text",
	}
	for _, marker := range required {
		if !strings.Contains(authorizeFormsSnapshotSQL, marker) {
			t.Fatalf("Forms snapshot authorize SQL omits %s", marker)
		}
	}
	for _, forbidden := range []string{
		"iam_users ", "iam_permissions ", "iam_authorization_clock ",
		"platform.", "public.", "pg_temp", "current_schema()", "$5",
	} {
		if strings.Contains(authorizeFormsSnapshotSQL, forbidden) {
			t.Fatalf("Forms snapshot authorize SQL contains forbidden token %q", forbidden)
		}
	}
}

func TestWorkflowHumanTaskAuthorizeSQLUsesExactV4EvidenceRoutine(t *testing.T) {
	required := []string{
		"iam_authorization_evidence_read_v4",
		"(pg_catalog.current_schemas(false))[2]::name",
		"$1::uuid", "$2::uuid", "$3::text", "$4::text",
	}
	for _, marker := range required {
		if !strings.Contains(authorizeWorkflowHumanTaskSQL, marker) {
			t.Fatalf("Workflow human task authorize SQL omits %s", marker)
		}
	}
	for _, forbidden := range []string{
		"iam_users ", "iam_permissions ", "iam_authorization_clock ",
		"platform.", "public.", "pg_temp", "current_schema()", "$5",
	} {
		if strings.Contains(authorizeWorkflowHumanTaskSQL, forbidden) {
			t.Fatalf("Workflow human task authorize SQL contains %q", forbidden)
		}
	}
}
