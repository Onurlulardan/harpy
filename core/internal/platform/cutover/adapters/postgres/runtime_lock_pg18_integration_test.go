//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCutoverRuntimeLocksAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startCutoverIntegrationDatabase(t, ctx)
	scopeID := uuid.MustParse("71000000-0000-4000-8000-000000000061")
	seedRuntimeLockAuthority(t, ctx, database.admin, scopeID)
	runtime := openRuntimeLockConnection(t, ctx, database.dsn)
	defer runtime.Close(ctx)
	transaction, err := runtime.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(ctx)
	var state, writer string
	var generation int64
	err = transaction.QueryRow(ctx, `SELECT state,writer,generation FROM
platform.cutover_runtime_lock_group_v1('tenant',$1,'workflow')`, scopeID).Scan(
		&state, &writer, &generation,
	)
	if err != nil || state != "v2" || writer != "v2" || generation != 7 {
		t.Fatalf("runtime group lock = %q %q %d: %v", state, writer, generation, err)
	}
	var member string
	err = transaction.QueryRow(ctx, `SELECT domain FROM
platform.cutover_runtime_lock_group_members_v1('tenant',$1,'workflow')`, scopeID).
		Scan(&member)
	if err != nil || member != "workflow" {
		t.Fatalf("runtime member lock = %q: %v", member, err)
	}
	err = transaction.QueryRow(ctx, `SELECT state,writer,generation FROM
platform.cutover_runtime_lock_domain_v1('tenant',$1,'workflow')`, scopeID).Scan(
		&state, &writer, &generation,
	)
	if err != nil || state != "v2" || writer != "v2" || generation != 7 {
		t.Fatalf("runtime domain lock = %q %q %d: %v", state, writer, generation, err)
	}
	assertRuntimeLockBlocksSwitch(t, ctx, database.admin, scopeID)
}

func seedRuntimeLockAuthority(
	t *testing.T, ctx context.Context, admin *pgx.Conn, scopeID uuid.UUID,
) {
	t.Helper()
	statements := []struct {
		document string
		args     []any
	}{
		{`INSERT INTO platform.domain_cutovers(
scope_kind,scope_id,domain,state,writer,generation,requested_by,reason,change_reference)
VALUES('tenant',$1,'workflow','v2','v2',7,$2,'runtime lock','P7-RUNTIME-LOCK')`,
			[]any{scopeID, integrationOwnerID}},
		{`INSERT INTO platform.cutover_groups(
scope_kind,scope_id,group_id,state,writer,generation,requested_by,reason,change_reference)
VALUES('tenant',$1,'workflow','v2','v2',7,$2,'runtime lock','P7-RUNTIME-LOCK')`,
			[]any{scopeID, integrationOwnerID}},
		{`INSERT INTO platform.cutover_group_members(
scope_kind,scope_id,group_id,ordinal,domain)
VALUES('tenant',$1,'workflow',1,'workflow')`, []any{scopeID}},
	}
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement.document, statement.args...); err != nil {
			t.Fatalf("seed runtime lock authority: %v", err)
		}
	}
}

func openRuntimeLockConnection(
	t *testing.T, ctx context.Context, dsn string,
) *pgx.Conn {
	t.Helper()
	configuration, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	configuration.User = "harpy_v2_runtime"
	configuration.Password = "integration-runtime-secret"
	configuration.RuntimeParams["role"] = "harpy_runtime"
	configuration.RuntimeParams["search_path"] = "pg_catalog"
	connection, err := pgx.ConnectConfig(ctx, configuration)
	if err != nil {
		t.Fatal(err)
	}
	return connection
}

func assertRuntimeLockBlocksSwitch(
	t *testing.T, ctx context.Context, admin *pgx.Conn, scopeID uuid.UUID,
) {
	t.Helper()
	transaction, err := admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(ctx)
	if _, err := transaction.Exec(ctx, "SET LOCAL lock_timeout='100ms'"); err != nil {
		t.Fatal(err)
	}
	_, err = transaction.Exec(ctx, `UPDATE platform.cutover_groups
SET reason=reason WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='workflow'`, scopeID)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "55P03" {
		t.Fatalf("concurrent switch error = %v, want lock timeout", err)
	}
}
