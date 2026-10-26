package postgres

import (
	"strings"
	"testing"
)

func TestReferenceSQLUsesProductionTenantScopeAndOwnedRoutine(t *testing.T) {
	if !strings.Contains(referenceEvidenceSQL,
		"(pg_catalog.current_schemas(false))[2]::name") ||
		!strings.Contains(referenceEvidenceSQL,
			"system_resource_layout_field_reference_evidence_v2") ||
		strings.Contains(referenceEvidenceSQL, "current_schema()") ||
		strings.Contains(referenceEvidenceSQL, " JOIN ") {
		t.Fatalf("unsafe layout reference SQL: %s", referenceEvidenceSQL)
	}
}
