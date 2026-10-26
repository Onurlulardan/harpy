package maintenance

import (
	"strings"
	"testing"
)

func TestMaintenanceReadsUseOnlyClaimBoundRoutines(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		routine   string
	}{
		{"directory repair", readDirectoryPageSQL,
			"iam_maintenance_directory_repair_page_v1"},
		{"directory checksum", readDirectorySourcesSQL,
			"iam_maintenance_directory_checksum_sources_v1"},
		{"profile page", readProfilePageSQL,
			"iam_maintenance_profile_page_read_v1"},
		{"profile mappings", readProfileMappingsSQL,
			"iam_maintenance_profile_mappings_read_v1"},
		{"profile schema", readProfileSchemaSQL,
			"iam_maintenance_profile_schema_read_v2"},
		{"profile resources", readProfileResourcesSQL,
			"iam_maintenance_profile_lineage_resources_read_v3"},
		{"profile evidence", readProfileResourceEvidenceSQL,
			"iam_maintenance_profile_lineage_resources_read_v3"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lower := strings.ToLower(test.statement)
			if !strings.Contains(lower, test.routine) ||
				!strings.Contains(lower, "pg_catalog.current_schema()") {
				t.Fatalf("read is not routed through %s", test.routine)
			}
			for _, forbidden := range []string{
				"from iam_users", "from iam_user_identifiers",
				"from iam_profile_", "join iam_profile_",
				"from iam_maintenance_checkpoints",
			} {
				if strings.Contains(lower, forbidden) {
					t.Fatalf("read contains direct table access %q", forbidden)
				}
			}
		})
	}
}
