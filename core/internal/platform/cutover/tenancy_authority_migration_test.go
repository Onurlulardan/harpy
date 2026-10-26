package cutover

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestTenancyCutoverAuthorityMigrationsArePaired(t *testing.T) {
	versions := []string{
		"20260802235005",
		"20260802235006",
		"20260802235007",
	}
	for _, version := range versions {
		pair := tenantSingletonAuthorityMigrationPair(t, version)
		if len(pair) != 2 || pair["fresh"] != pair["v1_adopted"] {
			t.Fatalf("%s tenancy cutover migration is not paired", version)
		}
		if strings.Contains(pair["fresh"], "-- +goose Down") {
			t.Fatalf("%s tenancy cutover migration has a down path", version)
		}
		if _, err := postgresv1.InspectMigrationSource([]byte(pair["fresh"])); err != nil {
			t.Fatalf("%s tenancy cutover migration is not Goose-canonical: %v", version, err)
		}
	}
}

func TestTenancyCutoverAuthoritySuccessorDelegatesToEnsure(t *testing.T) {
	ensure := tenantSingletonAuthorityMigrationPair(t, "20260802235005")["fresh"]
	backfill := tenantSingletonAuthorityMigrationPair(t, "20260802235006")["fresh"]
	wrapper := tenantSingletonAuthorityMigrationPair(t, "20260802235007")["fresh"]
	for _, marker := range []string{
		"CREATE FUNCTION platform.cutover_ensure_tenancy_authority_v1(",
		"ARRAY['tenancy']",
	} {
		if !strings.Contains(ensure, marker) {
			t.Fatalf("tenancy ensure migration lacks %q", marker)
		}
	}
	for _, marker := range []string{
		"DROP CONSTRAINT domain_cutovers_domain_check",
		"DROP CONSTRAINT cutover_group_members_domain_check",
		"DROP CONSTRAINT cutover_transition_events_domain_check",
		"DROP CONSTRAINT contract_retirements_domain_check",
		"platform.cutover_ensure_mcp_ai_authority_v1(",
		"platform.cutover_ensure_tenancy_authority_v1(",
		"'tenancy', 'work', 'workflow'",
	} {
		if !strings.Contains(backfill, marker) {
			t.Fatalf("tenancy backfill lacks %q", marker)
		}
	}
	for _, marker := range []string{
		"CREATE FUNCTION platform.cutover_initialize_tenant_v5(",
		"platform.cutover_ensure_tenancy_authority_v1(",
		"CREATE OR REPLACE FUNCTION platform.cutover_initialize_tenant_v3(",
		"platform.cutover_initialize_tenant_v5(",
	} {
		if !strings.Contains(wrapper, marker) {
			t.Fatalf("tenancy successor migration lacks %q", marker)
		}
	}
}
