//go:build integration

package migrator

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

var tenantTargetA = postgresstore.TenantMigrationTarget{
	TenantID: "aaaaaaaa-0000-4000-8000-000000000001", SchemaName: "tenant_a",
}
var tenantTargetB = postgresstore.TenantMigrationTarget{
	TenantID: "bbbbbbbb-0000-4000-8000-000000000002", SchemaName: "tenant_b",
}
var tenantTargetRace = postgresstore.TenantMigrationTarget{
	TenantID: "cccccccc-0000-4000-8000-000000000003", SchemaName: "tenant_race",
}

func TestTenantMigrationExecutorPostgreSQL184IsolationAndLockRace(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	for _, target := range []postgresstore.TenantMigrationTarget{
		tenantTargetA, tenantTargetB, tenantTargetRace,
	} {
		insertTenantTarget(t, ctx, fixture.database, target)
	}
	for _, target := range []postgresstore.TenantMigrationTarget{tenantTargetA, tenantTargetB} {
		applied, err := runner.executeTenantTarget(ctx, target, tenantRunOne)
		if err != nil || !reflect.DeepEqual(applied, tenantVersions(runner)) {
			t.Fatalf("apply %s: versions=%v err=%v", target.TenantID, applied, err)
		}
		assertTenantMigrationApplied(t, ctx, fixture.database, target, "new")
	}
	assertTenantIsolation(t, ctx, fixture.database)
	var wait sync.WaitGroup
	results := make(chan []int64, 2)
	errors := make(chan error, 2)
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			applied, err := runner.executeTenantTarget(ctx, tenantTargetRace, tenantRunOne)
			results <- applied
			errors <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("race executor error: %v", err)
		}
	}
	var appliedCount int
	for applied := range results {
		appliedCount += len(applied)
	}
	if appliedCount != len(runner.tenantSelected) {
		t.Fatalf("race applied version count = %d", appliedCount)
	}
	var attempts int
	if err := fixture.database.QueryRowContext(ctx, `SELECT count(*)
FROM platform.tenant_migration_checkpoint_attempts WHERE tenant_id=$1`,
		tenantTargetRace.TenantID).Scan(&attempts); err != nil || attempts != 3 {
		t.Fatalf("race attempt receipts = %d err=%v", attempts, err)
	}
	assertTenantMigrationApplied(t, ctx, fixture.database, tenantTargetRace, "new")
}

func tenantVersions(runner *Runner) []int64 {
	versions := make([]int64, 0, len(runner.tenantSelected))
	for _, selected := range runner.tenantSelected {
		versions = append(versions, selected.Version)
	}
	return versions
}
