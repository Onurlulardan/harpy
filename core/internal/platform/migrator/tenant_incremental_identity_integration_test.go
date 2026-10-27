//go:build integration

package migrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"testing/fstest"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func TestTenantMigrationPostgreSQL184IncrementalReleasePreservesIdentity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	target := postgresstore.TenantMigrationTarget{
		TenantID:   "17171717-0000-4000-8000-000000000013",
		SchemaName: "tenant_incremental",
	}
	insertTenantTarget(t, ctx, fixture.database, target)
	runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	if _, err := runner.executeTenantTarget(ctx, target, tenantRunOne); err != nil {
		t.Fatal(err)
	}
	var createdAt time.Time
	if err := fixture.database.QueryRowContext(ctx,
		`SELECT created_at FROM tenant_incremental.tenancy_schema_identity`,
	).Scan(&createdAt); err != nil {
		t.Fatal(err)
	}

	const filename = "20260725124200_incremental.fresh.sql"
	content := []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n")
	digest := sha256.Sum256(content)
	files := runner.tenantFiles.(fstest.MapFS)
	files[filename] = &fstest.MapFile{Data: content}
	runner.tenantSelected = append(runner.tenantSelected, postgresv1.SelectedMigration{
		Owner: "tenancy", Scope: "tenant", Version: 20260725124200,
		TransactionMode: "transactional", Phase: "expand",
		TargetDatabaseGeneration: 2, NextMinimumRuntimeGeneration: 1,
		Origin: postgresv1.DatabaseOriginFresh, Path: "migrations/tenant/" + filename,
		SHA256: hex.EncodeToString(digest[:]),
	})
	applied, err := runner.executeTenantTarget(ctx, target, tenantRunTwo)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 1 || applied[0] != 20260725124200 {
		t.Fatalf("incremental applied migrations = %v", applied)
	}
	var after time.Time
	if err := fixture.database.QueryRowContext(ctx,
		`SELECT created_at FROM tenant_incremental.tenancy_schema_identity`,
	).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if !after.Equal(createdAt) {
		t.Fatalf("tenant identity creation changed: before=%s after=%s", createdAt, after)
	}
}
