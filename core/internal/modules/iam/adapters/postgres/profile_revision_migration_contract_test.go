package postgres

import (
	"strings"
	"testing"
)

func TestIAMProfileRevisionMigrationContracts(t *testing.T) {
	tests := []struct {
		index   int
		markers []string
	}{
		{0, []string{"ADD COLUMN field_key text COLLATE \"C\"", "semantic_sha256 bytea"}},
		{1, []string{"CREATE FUNCTION iam_profile_validation_canonical_v1", "maximum_bytes"}},
		{2, []string{"CREATE FUNCTION iam_profile_semantic_sha256_v1", "\"key\":"}},
		{3, []string{"UPDATE iam_profile_field_revisions AS snapshot", "field_state = field.state"}},
		{4, []string{"ALTER COLUMN semantic_sha256 SET NOT NULL", "validation_digest_check"}},
		{5, []string{"CREATE FUNCTION iam_profile_definition_sha256_v1", "\"state\":"}},
		{6, []string{"UPDATE iam_profile_schema_versions", "definition_sha256"}},
		{7, []string{"iam_forbid_profile_destructive_mutation_v1", "BEFORE TRUNCATE"}},
		{8, []string{"iam_guard_profile_schema_lifecycle_v1", "draft", "active", "retired"}},
		{9, []string{"iam_guard_profile_field_lifecycle_v1", "must begin as active"}},
		{10, []string{"iam_guard_profile_snapshot_insert_v1", "FOR KEY SHARE"}},
		{11, []string{"iam_profile_historical_read_v1", "active', 'retired"}},
		{12, []string{"iam_profile_active_catalog_read_v1", "state = 'active'"}},
		{13, []string{"iam_profile_document_read_v1", "FROM iam_users"}},
		{14, []string{"GRANT EXECUTE", "TO harpy_runtime"}},
	}
	for _, test := range tests {
		name := iamProfileRevisionMigrationNames[test.index]
		t.Run(name, func(t *testing.T) {
			source := string(readIAMIdentityMigration(t, "fresh", name))
			for _, marker := range test.markers {
				if !strings.Contains(source, marker) {
					t.Fatalf("profile revision migration lacks %q", marker)
				}
			}
		})
	}
}

func TestIAMProfileRevisionUsesReservedVersions(t *testing.T) {
	want := []string{
		"20260716013200", "20260716013201", "20260716013202",
		"20260716013203", "20260716013204", "20260716013205",
		"20260716013206", "20260716013210", "20260716013211",
		"20260716013212", "20260716013213", "20260716013215",
		"20260716013216", "20260716013217", "20260716013220",
	}
	for index, name := range iamProfileRevisionMigrationNames {
		if !strings.HasPrefix(name, want[index]+"_") {
			t.Fatalf("profile revision migration %q is outside %s", name, want[index])
		}
	}
}
