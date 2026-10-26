//go:build integration

package migrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

const finalizedMigrationRun = "40000000-0000-4000-8000-000000000004"

type finalizedSchemaAuthority struct {
	value runtimev1.BinarySchemaContract
}

func (authority finalizedSchemaAuthority) Contract() runtimev1.BinarySchemaContract {
	return authority.value
}

func TestTenantMigrationPostgreSQL184FinalizedReleaseDeniesLateTenantWithoutWrites(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	runner, selected := finalizedReleaseRunner(t, ctx, fixture)
	seedFinalizedRelease(t, ctx, fixture, runner, selected)
	target := postgresstore.TenantMigrationTarget{
		TenantID: "16161616-0000-4000-8000-000000000012", SchemaName: "tenant_late",
	}
	insertTenantTarget(t, ctx, fixture.database, target)
	beforeRuns := countFinalizedRows(t, ctx, fixture, "platform.migration_runs")
	result, err := runner.Up(ctx, UpRequest{
		RunID: "40000000-0000-4000-8000-000000000005",
	})
	if err == nil || !strings.Contains(err.Error(), "use tenant onboarding") {
		t.Fatalf("late tenant finalized release result=%#v err=%v", result, err)
	}
	if after := countFinalizedRows(t, ctx, fixture, "platform.migration_runs"); after != beforeRuns {
		t.Fatalf("migration run count changed: before=%d after=%d", beforeRuns, after)
	}
	if countFinalizedRows(t, ctx, fixture, "platform.tenant_migration_checkpoints") != 0 ||
		countFinalizedRows(t, ctx, fixture, "platform.tenant_migration_checkpoint_attempts") != 0 {
		t.Fatal("late tenant denial wrote checkpoint state")
	}
	var relation *string
	if err := fixture.database.QueryRowContext(ctx,
		"SELECT pg_catalog.to_regclass($1)::text", target.SchemaName+".tenancy_schema_identity",
	).Scan(&relation); err != nil || relation != nil {
		t.Fatalf("late tenant relation = %v err=%v", relation, err)
	}
}

func finalizedReleaseRunner(
	t *testing.T, ctx context.Context, fixture tenantMigrationFixture,
) (*Runner, postgresv1.SelectedMigration) {
	t.Helper()
	connection, err := fixture.database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	live, err := postgresstore.ReadLiveDatabaseIdentity(ctx, connection)
	_ = connection.Close()
	if err != nil {
		t.Fatal(err)
	}
	const filename = "20260714230000_finalized.fresh.sql"
	content := []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n")
	digest := sha256.Sum256(content)
	selected := postgresv1.SelectedMigration{
		Owner: "platform_migrator", Scope: "platform", Version: 20260714230000,
		TransactionMode: "transactional", Phase: "expand",
		TargetDatabaseGeneration: 2, NextMinimumRuntimeGeneration: 1,
		Origin: postgresv1.DatabaseOriginFresh,
		Path:   "migrations/" + filename, SHA256: hex.EncodeToString(digest[:]),
	}
	tenant := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	manifest, extension := strings.Repeat("a", 64), strings.Repeat("b", 64)
	contract := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion, Component: "core",
		RuntimeGeneration: 2, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: manifest, ExtensionManifestSHA256: extension,
		PostgresImageDigest: runtimev1.PostgresImageDigest,
	}
	return &Runner{
		database: fixture.database, platformFiles: fstest.MapFS{
			filename: &fstest.MapFile{Data: content},
		},
		selected:       []postgresv1.SelectedMigration{selected},
		tenantDatabase: tenant.tenantDatabase, tenantFiles: tenant.tenantFiles,
		tenantSelected: tenant.tenantSelected,
		identity: runIdentity{
			releaseID:      "aaaaaaaa-1111-4111-8111-111111111111",
			environmentID:  "aaaaaaaa-2222-4222-8222-222222222222",
			databaseName:   live.Name,
			databaseOrigin: postgresv1.DatabaseOriginFresh,
			databaseOwner:  live.Owner, databaseEncoding: live.Encoding,
			databaseLocaleProvider: live.LocaleProvider,
			databaseCollation:      live.Collation, databaseCType: live.CType,
			dataChecksums:  live.Checksums,
			manifestSHA256: manifest, extensionSHA256: extension,
			authoritySHA256: strings.Repeat("c", 64), databaseSystemID: live.SystemID,
			databaseOID: live.OID, databaseGeneration: 2, minimumRuntimeGeneration: 1,
		},
		evidence:        foundationEvidence{receipts: finalizedReceipts()},
		schemaAuthority: finalizedSchemaAuthority{value: contract},
		cleanupTimeout:  requiredCleanupTimeout,
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
	}, selected
}

func finalizedReceipts() []receiptEvidence {
	kinds := []string{"cluster_bootstrap", "extension_admin", "credential_cleanup"}
	receipts := make([]receiptEvidence, 0, len(kinds))
	for index, kind := range kinds {
		receipts = append(receipts, receiptEvidence{
			kind: kind, runID: "40000000-0000-4000-8000-00000000000" + string(rune('1'+index)),
		})
	}
	return receipts
}

func countFinalizedRows(
	t *testing.T, ctx context.Context, fixture tenantMigrationFixture, table string,
) int {
	t.Helper()
	var count int
	if err := fixture.database.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
