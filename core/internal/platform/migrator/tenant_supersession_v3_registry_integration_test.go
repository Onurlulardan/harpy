//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"
)

func assertTenantSupersessionV3Registry(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	var distribution string
	err := database.QueryRowContext(ctx, `SELECT pg_catalog.string_agg(
 distance::text||':'||edge_count::text,',' ORDER BY distance)
FROM (SELECT supersession_distance AS distance,pg_catalog.count(*) AS edge_count
FROM platform.tenant_migration_supersessions
WHERE owner='system_resource_layout' AND scope='tenant'
  AND logical_version=20260730120104
GROUP BY supersession_distance) edges`).Scan(&distribution)
	if err != nil || distribution != "1:6,2:4,3:2" {
		t.Fatalf("v3 registry distribution=%q: %v", distribution, err)
	}
}

func assertSucceededAliasesSatisfyTerminal(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	aliases := []struct{ tenantID, path, sha string }{
		{"018f0000-0000-7000-8000-000000000313", chainOriginal, chainOriginalSH},
		{"018f0000-0000-7000-8000-000000000314", chainFirst, chainFirstSH},
		{"018f0000-0000-7000-8000-000000000315", chainSecond, chainSecondSH},
	}
	setChainCheckpointTrigger(t, ctx, database, false)
	defer setChainCheckpointTrigger(t, ctx, database, true)
	for _, alias := range aliases {
		_, err := database.ExecContext(ctx, `INSERT INTO
platform.tenant_migration_checkpoints (
 tenant_id,owner,logical_version,database_origin,selected_path,
 selected_sha256,run_id,state,attempt,started_at,finished_at
) VALUES ($1,'system_resource_layout',$2,'fresh',$3,$4,$5,'succeeded',1,
 pg_catalog.clock_timestamp()-INTERVAL '2 seconds',pg_catalog.clock_timestamp())`,
			alias.tenantID, chainVersion, alias.path, alias.sha, chainOldRunID)
		if err != nil {
			t.Fatal(err)
		}
		var matches bool
		err = database.QueryRowContext(ctx, `SELECT
platform.platform_migrator_tenant_checkpoint_satisfies_v3($1,
'system_resource_layout',$2,'fresh',$3,$4)`, alias.tenantID,
			chainVersion, chainTerminal, chainTerminalSH).Scan(&matches)
		if err != nil || !matches {
			t.Fatalf("succeeded alias %s did not satisfy terminal: %t %v",
				alias.path, matches, err)
		}
	}
}
