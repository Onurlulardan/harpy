//go:build integration

package migrator

import (
	"context"
	"reflect"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func TestTenantMigrationExecutorPostgreSQL184UsesCanonicalUUIDByteOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	low := postgresstore.TenantMigrationTarget{
		TenantID: "00000000-0000-4000-8000-000000000010", SchemaName: "tenant_order_low",
	}
	high := postgresstore.TenantMigrationTarget{
		TenantID: "ffffffff-ffff-4fff-bfff-fffffffffff0", SchemaName: "tenant_order_high",
	}
	insertTenantTarget(t, ctx, fixture.database, high)
	insertTenantTarget(t, ctx, fixture.database, low)
	connection, err := fixture.database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	targets, err := postgresstore.ListTenantMigrationTargets(ctx, connection)
	if err != nil || !reflect.DeepEqual(targets, []postgresstore.TenantMigrationTarget{low, high}) {
		t.Fatalf("canonical target order = %#v err=%v", targets, err)
	}
	applied, err := runner.runTenantMigrations(ctx, connection, tenantRunOne)
	if err != nil || !reflect.DeepEqual(applied, tenantVersions(runner)) {
		t.Fatalf("ordered tenant execution = %v err=%v", applied, err)
	}
	var lowReceipt, highReceipt time.Time
	query := `SELECT min(recorded_at) FROM platform.tenant_migration_checkpoint_attempts
WHERE tenant_id=$1`
	if err := fixture.database.QueryRowContext(ctx, query, low.TenantID).Scan(&lowReceipt); err != nil {
		t.Fatal(err)
	}
	if err := fixture.database.QueryRowContext(ctx, query, high.TenantID).Scan(&highReceipt); err != nil {
		t.Fatal(err)
	}
	if !lowReceipt.Before(highReceipt) {
		t.Fatalf("execution order timestamps = low %s high %s", lowReceipt, highReceipt)
	}
}
