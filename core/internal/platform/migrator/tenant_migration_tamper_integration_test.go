//go:build integration

package migrator

import (
	"context"
	"strings"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/jackc/pgx/v5"
)

func TestTenantMigrationPostgreSQL184TamperAndReconnectFailClosed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	selected := runner.tenantSelected[0]
	cases := []struct {
		name, id, schema string
		mutate           func(*postgresv1.SelectedMigration)
	}{
		{"path", "12121212-0000-4000-8000-000000000007", "tenant_path_tamper",
			func(value *postgresv1.SelectedMigration) { value.Path = "migrations/tenant/tampered.sql" }},
		{"origin", "12121212-0000-4000-8000-000000000008", "tenant_origin_tamper",
			func(value *postgresv1.SelectedMigration) { value.Origin = postgresv1.DatabaseOriginV1Adopted }},
		{"hash", "12121212-0000-4000-8000-000000000009", "tenant_hash_tamper",
			func(value *postgresv1.SelectedMigration) { value.SHA256 = strings.Repeat("b", 64) }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			tampered := postgresstore.TenantMigrationTarget{
				TenantID: testCase.id, SchemaName: testCase.schema,
			}
			insertTenantTarget(t, ctx, fixture.database, tampered)
			checkpoint := selected
			testCase.mutate(&checkpoint)
			_, err := fixture.database.ExecContext(ctx, `INSERT INTO platform.tenant_migration_checkpoints
(tenant_id,owner,logical_version,database_origin,selected_path,selected_sha256,
run_id,state,attempt) VALUES($1,$2,$3,$4,$5,$6,$7,'pending',1)`,
				tampered.TenantID, checkpoint.Owner, checkpoint.Version, checkpoint.Origin,
				checkpoint.Path, checkpoint.SHA256, tenantRunOne)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := runner.executeTenantTarget(ctx, tampered, tenantRunOne); err == nil ||
				!strings.Contains(err.Error(), "differs from verified selection") {
				t.Fatalf("tampered checkpoint error = %v", err)
			}
			var relation *string
			qualified := tampered.SchemaName + ".tenancy_schema_identity"
			if err := fixture.database.QueryRowContext(
				ctx, "SELECT pg_catalog.to_regclass($1)::text", qualified,
			).Scan(&relation); err != nil || relation != nil {
				t.Fatalf("tampered target relation = %v err=%v", relation, err)
			}
		})
	}
	reconnect := postgresstore.TenantMigrationTarget{
		TenantID: "13131313-0000-4000-8000-000000000008", SchemaName: "tenant_reconnect",
	}
	insertTenantTarget(t, ctx, fixture.database, reconnect)
	database := openOneShotTenantDatabase(fixture.config.Copy())
	defer database.Close()
	if err := postgresstore.AcquireTenantMigrationLock(ctx, database, reconnect.TenantID); err != nil {
		t.Fatal(err)
	}
	searchPath := pgx.Identifier{reconnect.SchemaName}.Sanitize() + ",pg_catalog"
	if err := postgresstore.SetTenantMigrationSearchPath(ctx, database, searchPath); err != nil {
		t.Fatal(err)
	}
	proof, err := postgresstore.ReadTenantMigrationSession(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	selected = runner.tenantSelected[0]
	started, err := prepareTenantCheckpoint(
		ctx, database, reconnect, proof, selected, runner.tenantAliases, tenantRunOne,
	)
	if err != nil || !started {
		t.Fatalf("prepare ambiguous checkpoint: started=%v err=%v", started, err)
	}
	terminateTenantMigrationBackend(t, ctx, fixture.config, proof.BackendPID)
	_, _, err = reconcileTenantMigration(
		ctx, database, reconnect, proof, selected, runner.tenantAliases,
	)
	if err == nil {
		t.Fatalf("ambiguous one-shot reconcile error = %v", err)
	}
	var state string
	var attempts int
	if err := fixture.database.QueryRowContext(ctx, `SELECT count(*)
FROM platform.tenant_migration_checkpoint_attempts WHERE tenant_id=$1`,
		reconnect.TenantID).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if err := fixture.database.QueryRowContext(ctx, `SELECT state
FROM platform.tenant_migration_checkpoints WHERE tenant_id=$1 AND logical_version=$2`,
		reconnect.TenantID, selected.Version).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "running" || attempts != 0 {
		t.Fatalf("ambiguous checkpoint = state %s attempts %d", state, attempts)
	}
}

func terminateTenantMigrationBackend(
	t *testing.T,
	ctx context.Context,
	config *pgx.ConnConfig,
	backendPID int64,
) {
	t.Helper()
	connection, err := pgx.ConnectConfig(ctx, config.Copy())
	if err != nil {
		t.Fatalf("open dedicated backend terminator: %v", err)
	}
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	defer func() {
		if err := connection.Close(cleanup); err != nil {
			t.Errorf("close dedicated backend terminator: %v", err)
		}
	}()
	if _, err := connection.Exec(ctx, "SET ROLE pg_signal_backend"); err != nil {
		t.Fatalf("assume backend terminator role: %v", err)
	}
	var terminated bool
	terminateErr := connection.QueryRow(
		ctx, "SELECT pg_catalog.pg_terminate_backend($1,5000)", backendPID,
	).Scan(&terminated)
	_, resetErr := connection.Exec(cleanup, "RESET ROLE")
	if terminateErr != nil || !terminated {
		t.Fatalf("terminate tenant backend = %v err=%v", terminated, terminateErr)
	}
	if resetErr != nil {
		t.Fatalf("reset backend terminator role: %v", resetErr)
	}
}
