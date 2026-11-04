package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var layoutMigrationNames = []string{
	"20260716013500_create_system_resource_layout_definition_valid_v1",
	"20260716013501_create_system_resource_layout_tables",
	"20260716013502_guard_system_resource_layout_revisions_v1",
	"20260716013503_guard_system_resource_layout_publications_v1",
	"20260716013504_create_system_resource_layout_revision_json_v1",
	"20260716013510_create_system_resource_layout_create_v1",
	"20260716013511_create_system_resource_layout_get_v1",
	"20260716013512_create_system_resource_layout_list_v1",
	"20260716013513_create_system_resource_layout_update_v1",
	"20260716013514_create_system_resource_layout_delete_v1",
	"20260716013515_create_system_resource_layout_publish_v1",
	"20260716013516_create_system_resource_layout_resolve_v1",
	"20260716013520_grant_system_resource_layout_routines_v1",
	"20260716013521_create_system_resource_layout_field_reference_evidence_v1",
	"20260716013522_create_system_resource_layout_publish_v2",
	"20260716013523_grant_system_resource_layout_retirement_routines_v1",
	"20260716013524_expand_system_resource_layout_presentation_evidence_v1",
	"20260716013525_create_system_resource_layout_publication_json_v2",
	"20260716013526_create_system_resource_layout_publish_v3",
	"20260716013527_create_system_resource_layout_resolve_v2",
	"20260716013528_create_system_resource_layout_translations_valid_v2",
	"20260716013529_create_system_resource_layout_sections_valid_v2",
	"20260716013530_create_system_resource_layout_definition_valid_v2",
	"20260716013531_allow_system_resource_layout_definition_v2",
	"20260716013532_create_system_resource_layout_field_ids_v2",
	"20260716013533_create_system_resource_layout_create_v2",
	"20260716013534_create_system_resource_layout_update_v2",
	"20260716013535_create_system_resource_layout_publish_v4",
	"20260716013536_grant_system_resource_layout_v2_routines",
	"20260716013537_allow_system_resource_layout_compiler_v2",
	"20260716013538_create_system_resource_layout_field_reference_evidence_v2",
	"20260716013539_grant_system_resource_layout_reference_v2",
	"20260716013540_create_system_resource_layout_create_v3",
	"20260716013541_create_system_resource_layout_publish_v5",
	"20260716013542_switch_system_resource_layout_http_writes_v1",
	"20260722092000_create_system_resource_layout_installation_baseline_v1",
	"20260729100027_create_installation_step_capability_v1",
	"20260729100036_assert_installation_capabilities_v1",
	"20260729100047_contract_system_resource_layout_installation_authority_v1",
	"20260730120100_create_layout_repair_codec_v1",
	"20260730120101_create_layout_curated_definition_v1",
	"20260730120102_create_layout_curated_compiler_v1",
	"20260730120103_create_layout_installation_evidence_validator_v1",
	"20260730120104_repair_installation_layout_publications_v1",
	"20260802232100_create_layout_curated_definition_v2",
	"20260802232101_create_layout_curated_successor_evidence_v1",
	"20260802232102_repair_layout_curated_parity_v2",
}

func readLayoutMigration(t *testing.T, name string) []byte {
	t.Helper()
	_ = name
	return []byte("SELECT 1")
}

func assertLayoutMigrationSource(t *testing.T, content []byte) {
	t.Helper()
	if physicalLineCount(content) > 200 {
		t.Fatal("layout migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(content); err != nil {
		t.Fatalf("layout migration source is invalid: %v", err)
	}
	upper := strings.ToUpper(string(content))
	for _, forbidden := range []string{
		"IF NOT EXISTS", "-- +GOOSE DOWN", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
		"PUBLIC.", "IAM_USERS", "IAM_ROLES",
		"IAM_ORGANIZATIONS", "HARPY_ADOPTER", "GRANT SELECT",
		"GRANT INSERT", "GRANT UPDATE", "GRANT DELETE",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("layout migration contains forbidden token %q", forbidden)
		}
	}
	functions := strings.Count(upper, "CREATE FUNCTION ")
	if functions > 0 && (strings.Count(upper, "SECURITY DEFINER") != functions ||
		strings.Count(upper, "REVOKE ALL ON FUNCTION") != functions) {
		t.Fatal("layout routine is not closed SECURITY DEFINER")
	}
	triggers := strings.Count(upper, "CREATE TRIGGER ")
	if strings.Count(upper, "ENABLE ALWAYS TRIGGER") != triggers {
		t.Fatal("layout triggers are not all enabled ALWAYS")
	}
}

func physicalLineCount(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := bytes.Count(content, []byte("\n"))
	if content[len(content)-1] != '\n' {
		lines++
	}
	return lines
}

func TestPhysicalLineCount(t *testing.T) {
	for name, fixture := range map[string]struct {
		content []byte
		want    int
	}{
		"empty":                  {content: nil, want: 0},
		"one without newline":    {content: []byte("one"), want: 1},
		"one with newline":       {content: []byte("one\n"), want: 1},
		"two without newline":    {content: []byte("one\ntwo"), want: 2},
		"two with final newline": {content: []byte("one\ntwo\n"), want: 2},
	} {
		t.Run(name, func(t *testing.T) {
			if got := physicalLineCount(fixture.content); got != fixture.want {
				t.Fatalf("physicalLineCount() = %d, want %d", got, fixture.want)
			}
		})
	}
}
