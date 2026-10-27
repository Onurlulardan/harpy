package migrator

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var tenantSupersessionChainMigrations = []string{
	"20260730120305_expand_tenant_migration_supersession_chain_v2",
	"20260730120306_allow_checkpoint_supersession_chain_v2",
	"20260730120307_create_tenant_checkpoint_requirement_matcher_v2",
	"20260730120308_create_tenant_requirement_verifier_v4",
	"20260730120309_assert_tenant_supersession_chain_release_frontier_v2",
	"20260730120310_expand_tenant_migration_supersession_chain_v3",
	"20260730120311_allow_checkpoint_supersession_chain_v3",
	"20260730120312_create_tenant_checkpoint_requirement_matcher_v3",
	"20260730120313_create_tenant_requirement_verifier_v5",
	"20260730120314_assert_tenant_supersession_release_frontier_v3",
}

func TestTenantSupersessionChainMigrationsArePairedAndClosed(t *testing.T) {
	var combined []byte
	for _, name := range tenantSupersessionChainMigrations {
		fresh := readIAMNamedPlatformMigration(t, "fresh", name)
		if len(fresh) == 0 {
			t.Fatalf("fresh tenant supersession chain migration did not load: %s", name)
		}
		if bytes.Count(fresh, []byte("\n"))+1 > 200 {
			t.Fatalf("tenant supersession chain migration exceeds 200 lines: %s", name)
		}
		if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
			t.Fatalf("inspect tenant supersession chain migration %s: %v", name, err)
		}
		combined = append(combined, fresh...)
	}
	source := string(combined)
	for _, marker := range []string{
		"tenant_migration_supersessions_edge_unique UNIQUE",
		"database_origin,superseded_path,current_path",
		"supersession_distance IN (1,2)",
		"/supersessions/v2/fresh/",
		"/supersessions/v2/v1_adopted/",
		"platform_migrator_validate_checkpoint_transition_v4",
		"platform_migrator_tenant_checkpoint_satisfies_v2",
		"platform_migrator_verify_tenant_requirement_catalog_v4",
		"tenant supersession chain hardening advances the release frontier to 82",
		"pg_catalog.count(*) = 6",
		"supersession_distance IN (1,2,3)",
		"/supersessions/v3/fresh/",
		"/supersessions/v3/v1_adopted/",
		"platform_migrator_validate_checkpoint_transition_v5",
		"platform_migrator_tenant_checkpoint_satisfies_v3",
		"platform_migrator_verify_tenant_requirement_catalog_v5",
		"tenant supersession v3 hardening advances the release frontier to 83",
		"pg_catalog.count(*) = 12",
	} {
		if !strings.Contains(strings.ToLower(source), strings.ToLower(marker)) {
			t.Fatalf("tenant supersession chain lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"-- +goose Down", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "pg_temp",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("tenant supersession chain contains %q", forbidden)
		}
	}
}
