//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func prepareIAMSystemResourcesAsyncSurfaces(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	origin string,
) {
	t.Helper()
	migrator := openCutoverMigrator(t, ctx, database.dsn)
	defer migrator.Close(ctx)
	if _, err := migrator.Exec(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatalf("set async fixture owner: %v", err)
	}
	if _, err := migrator.Exec(ctx,
		"CREATE TABLE platform.jobs (id uuid PRIMARY KEY)"); err != nil {
		t.Fatalf("create exact async job dependency: %v", err)
	}
	versions := []string{
		"20260715021510_create_v1_async_adoption_inventory",
		"20260715021520_create_v1_async_adoption_receipts",
	}
	for _, version := range versions {
		name := fmt.Sprintf("%s.%s.sql", version, origin)
		path := filepath.Join("..", "..", "..", "async", "migrations", origin, name)
		document, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read async adoption migration %s: %v", name, err)
		}
		if _, err := migrator.Exec(ctx, string(document)); err != nil {
			t.Fatalf("apply async adoption migration %s: %v", name, err)
		}
	}
}

func seedIAMSystemResourcesAcceptedZeroInventory(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	scopeID, domain, acceptanceSHA256 string,
) {
	t.Helper()
	inventoryMarker, componentMarker, receiptMarker := byte('e'), byte('a'), byte('8')
	if domain == "system_resource_layout" {
		inventoryMarker, componentMarker, receiptMarker = 'f', 'b', '9'
	} else if domain != "iam" {
		t.Fatalf("unsupported IAM system-resource async domain %q", domain)
	}
	_, err := database.admin.Exec(ctx, `WITH stamp AS (
    SELECT pg_catalog.clock_timestamp() AS now
), inserted AS (
    INSERT INTO platform.async_v1_adoption_runs (
        run_id,scope_kind,scope_id,domain,group_id,source_generation,
        target_generation,source_group_generation,target_group_generation,
        inventory_contract_version,inventory_revision,inventory_sha256,
        component_manifest_sha256,expected_enqueue_fences,
        expected_scheduler_fences,expected_consumer_fences,item_count,state,
        captured_at,created_at,accepted_at,acceptance_sha256
    ) SELECT pg_catalog.gen_random_uuid(),'tenant',$1,$2,'iam_system_resources',
        1,2,1,2,'async-adoption-inventory.v1',1,$3,$4,1,1,1,0,'accepted',
        stamp.now-pg_catalog.make_interval(secs=>2),
        stamp.now-pg_catalog.make_interval(secs=>1),stamp.now,$5
    FROM stamp RETURNING run_id,inventory_sha256
)
INSERT INTO platform.async_v1_zero_inventory_receipts (
    run_id,inventory_contract_version,inventory_sha256,queue_count,
    timer_count,future_count,running_count,observed_at,receipt_sha256,recorded_at
)
SELECT inserted.run_id,'async-adoption-zero-inventory.v1',
    inserted.inventory_sha256,0,0,0,0,stamp.now,$6,stamp.now
FROM inserted CROSS JOIN stamp`, scopeID, domain, digestMarker(inventoryMarker),
		digestMarker(componentMarker), acceptanceSHA256, digestMarker(receiptMarker))
	if err != nil {
		t.Fatalf("seed accepted zero async inventory for %s: %v", domain, err)
	}
}
