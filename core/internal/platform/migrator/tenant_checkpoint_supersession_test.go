package migrator

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func TestTenantCheckpointMatchesEverySupersessionChainIdentity(t *testing.T) {
	selected := postgresv1.SelectedMigration{
		Owner: "system_resource_layout", Scope: tenantMigrationScope,
		Version: 20260730120104, Origin: postgresv1.DatabaseOriginFresh,
		Path:   "migrations/supersessions/v2/fresh/repair.fresh.sql",
		SHA256: strings.Repeat("c", 64),
	}
	expected := postgresstore.TenantCheckpointIdentity{
		TenantID: "018f0000-0000-7000-8000-000000000001",
		Owner:    selected.Owner, Version: selected.Version,
		Origin: string(selected.Origin), SelectedPath: selected.Path,
		SelectedHash: selected.SHA256,
	}
	aliases := map[int64]historySupersession{
		selected.Version: {
			version: selected.Version, owner: selected.Owner, scope: selected.Scope,
			origin: selected.Origin, currentPath: selected.Path,
			currentSHA256: selected.SHA256,
			superseded: []historyIdentity{
				{path: "migrations/fresh/repair.fresh.sql", sha256: strings.Repeat("a", 64)},
				{path: "migrations/supersessions/fresh/repair.fresh.sql",
					sha256: strings.Repeat("b", 64)},
			},
		},
	}
	for _, identity := range aliases[selected.Version].superseded {
		checkpoint := postgresstore.TenantCheckpoint{
			TenantCheckpointIdentity: expected,
		}
		checkpoint.SelectedPath, checkpoint.SelectedHash = identity.path, identity.sha256
		if !tenantCheckpointMatchesVerifiedSelection(
			checkpoint, expected, selected, aliases,
		) {
			t.Fatalf("checkpoint chain identity was rejected: %s", identity.path)
		}
		checkpoint.SelectedHash = strings.Repeat("d", 64)
		if tenantCheckpointMatchesVerifiedSelection(
			checkpoint, expected, selected, aliases,
		) {
			t.Fatalf("tampered checkpoint chain identity was accepted: %s", identity.path)
		}
	}
}
