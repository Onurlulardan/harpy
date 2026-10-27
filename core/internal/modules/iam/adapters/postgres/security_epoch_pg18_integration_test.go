//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var securityEpochPlatformMigrations = []string{
	"20260722022000_create_iam_security_epochs",
	"20260722022001_guard_iam_security_epochs",
	"20260722022002_create_iam_security_epoch_initialize_v1",
	"20260723131000_create_iam_security_epoch_for_tenant_v1",
	"20260725123500_grant_iam_installation_security_epoch_runtime",
}

var securityEpochTenantMigrations = []string{
	"20260722022100_create_iam_security_epoch_current_v1",
	"20260722022101_bind_iam_credentials_security_epoch",
	"20260722022102_guard_iam_credential_security_epoch",
	"20260722022103_create_iam_session_issue_v2",
	"20260722022104_create_iam_refresh_credential_v3",
	"20260722022105_create_iam_refresh_rotate_v4",
	"20260722022106_create_iam_current_session_read_v2",
	"20260722022107_grant_iam_security_epoch_runtime_v1",
	"20260722022108_create_iam_security_epoch_assert_v1",
}

func TestIAMSecurityEpochAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "security_epoch_"+origin)
			applySecurityEpochFixture(t, ctx, database, origin)
			assertSecurityEpochBehavior(t, ctx, database)
		})
	}
}

func applySecurityEpochFixture(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, securityEpochPG18Fixture); err != nil {
		t.Fatalf("prepare security epoch fixture: %v", err)
	}
	applySecurityEpochFiles(t, ctx, database, "platform", origin,
		securityEpochPlatformMigrations, "")
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx,
		"SET search_path = t_epoch, pg_catalog; "+securityEpochPG18RoutineFixture,
	); err != nil {
		t.Fatalf("prepare security epoch routines: %v", err)
	}
	applySecurityEpochFiles(t, ctx, connection, "tenant", origin,
		securityEpochTenantMigrations, "t_epoch, pg_catalog")
}

type securityEpochExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func applySecurityEpochFiles(
	t *testing.T, ctx context.Context, executor securityEpochExecutor,
	scope, origin string, names []string, searchPath string,
) {
	t.Helper()
	for _, name := range names {
		path := filepath.Join("..", "..", "migrations", scope, origin,
			name+"."+origin+".sql")
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		statement := string(body)
		if searchPath != "" {
			statement = "SET search_path = " + searchPath + "; " + statement
		}
		if _, err = executor.ExecContext(ctx, statement); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}
}

func assertSecurityEpochBehavior(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime;
SET LOCAL ROLE harpy_runtime;
SET LOCAL application_name = 'harpy-v2:core-operator-initialize'`); err != nil {
		t.Fatalf("bind installation runtime: %v", err)
	}
	var initialized bool
	if err = tx.QueryRowContext(ctx, `SELECT platform.iam_security_epoch_initialize_v1(
'30000000-0000-4000-8000-000000000003', statement_timestamp())`).Scan(&initialized); err != nil || !initialized {
		t.Fatalf("initialize security epoch as runtime: initialized=%t err=%v", initialized, err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	var initial []byte
	if err := db.QueryRowContext(ctx,
		"SELECT security_epoch FROM platform.iam_security_epochs",
	).Scan(&initial); err != nil || len(initial) != 32 {
		t.Fatalf("read initial security epoch: len=%d err=%v", len(initial), err)
	}
	const sessionID = "41000000-0000-4000-8000-000000000001"
	const userID = "42000000-0000-4000-8000-000000000001"
	_, err = db.ExecContext(ctx, `
INSERT INTO t_epoch.iam_sessions(id,user_id,state,revision,updated_at,retain_until)
VALUES($1,$2,'active',1,transaction_timestamp(),transaction_timestamp()+interval '30 days')`,
		sessionID, userID)
	if err != nil {
		t.Fatalf("insert epoch-bound session: %v", err)
	}
	var rows int
	if err = db.QueryRowContext(ctx, `
SELECT count(*) FROM t_epoch.iam_current_session_read_v2('t_epoch',$1,$2)`,
		userID, sessionID).Scan(&rows); err != nil || rows != 1 {
		t.Fatalf("read epoch-bound session: rows=%d err=%v", rows, err)
	}
	_, err = db.ExecContext(ctx, `UPDATE platform.iam_security_epochs
SET security_epoch=sha256(security_epoch), generation=generation+1,
    rotated_at=rotated_at+interval '1 second'`)
	if err != nil {
		t.Fatalf("rotate security epoch: %v", err)
	}
	if _, err = db.ExecContext(ctx, `
UPDATE t_epoch.iam_sessions SET revision=revision+1 WHERE id=$1`, sessionID); err == nil {
		t.Fatal("stale active session accepted a non-terminal update")
	}
	if _, err = db.ExecContext(ctx, `
UPDATE t_epoch.iam_sessions SET state='revoked',revision=revision+1,
updated_at=transaction_timestamp(),revoked_at=transaction_timestamp() WHERE id=$1`,
		sessionID); err != nil {
		t.Fatalf("stale session could not be terminalized: %v", err)
	}
}
