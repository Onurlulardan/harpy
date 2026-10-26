package userrelationpostgres

import (
	"strings"
	"testing"
)

func TestReplaceDelegatesExactUserRelationRoutine(t *testing.T) {
	required := []string{
		"FROM iam_identity_user_relations_replace_v1(",
		"(pg_catalog.current_schemas(false))[2]::name",
		"$6::boolean, $7::uuid, $8::bytea, $9::bytea",
		"$10::boolean, $11::uuid[], $12::uuid[]",
		"$13::boolean, $14::uuid[], $15::uuid[]",
		"$16::uuid, $17::uuid, $18::bigint, $19::bigint",
	}
	for _, fragment := range required {
		if !strings.Contains(replaceUserRelationsSQL, fragment) {
			t.Fatalf("replace SQL is missing %q", fragment)
		}
	}
}
