package postgres

import (
	"strings"
	"testing"
)

func TestTargetApplySchemaBindingUsesExactSignedCatalogIdentity(t *testing.T) {
	required := []string{
		"namespace.oid = $1::oid",
		"namespace.nspname = $2::text",
		"relation.oid = $3::oid",
		"relation.relname = $4::text",
		"procedure.oid = $6::oid",
		"namespace.oid = procedure.pronamespace",
		"procedure.pronamespace = target.schema_oid::oid",
		"pg_catalog.to_regprocedure($5::text)::oid = procedure.oid",
		"pg_catalog.set_config(",
		"true",
	}
	for _, fragment := range required {
		if !strings.Contains(bindTargetApplySchemaSQL, fragment) {
			t.Fatalf("target apply schema binding omits %q", fragment)
		}
	}
	for _, forbidden := range []string{"::name", "EXECUTE ", "SET search_path", "SET LOCAL"} {
		if strings.Contains(bindTargetApplySchemaSQL, forbidden) {
			t.Fatalf("target apply schema binding contains %q", forbidden)
		}
	}
}
