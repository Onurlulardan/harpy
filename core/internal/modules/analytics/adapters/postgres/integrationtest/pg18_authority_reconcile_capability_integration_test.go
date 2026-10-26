//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func prepareAnalyticsAuthorityReconcileCapability(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const statement = `
CREATE TABLE platform.test_tenancy_authority_reconcile_scopes (
 backend_pid integer NOT NULL,
 transaction_id xid8 NOT NULL,
 owner_key text NOT NULL,
 PRIMARY KEY(backend_pid,transaction_id,owner_key)
);
CREATE FUNCTION platform.tenancy_authority_reconcile_scope_assert_v1(
 p_owner_key text
) RETURNS boolean LANGUAGE sql STABLE SECURITY DEFINER
SET search_path TO pg_catalog
RETURN SESSION_USER='harpy_v2_runtime'
 AND CURRENT_USER='harpy_owner'
 AND EXISTS(
  SELECT 1 FROM platform.test_tenancy_authority_reconcile_scopes AS scope
  WHERE scope.backend_pid=pg_catalog.pg_backend_pid()
    AND scope.transaction_id=pg_catalog.pg_current_xact_id()
    AND scope.owner_key=p_owner_key
 );
REVOKE ALL ON FUNCTION
 platform.tenancy_authority_reconcile_scope_assert_v1(text) FROM PUBLIC;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare Analytics reconcile capability: %v", err)
	}
}

func configureAnalyticsReconcileTransaction(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	authorized bool,
) {
	t.Helper()
	if authorized {
		if _, err := transaction.ExecContext(ctx, `
INSERT INTO platform.test_tenancy_authority_reconcile_scopes(
 backend_pid,transaction_id,owner_key
) VALUES(pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),'analytics')`,
		); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := transaction.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_owner;
SELECT pg_catalog.set_config(
 'application_name','harpy-v2:core-operator-initialize',true
)`); err != nil {
		t.Fatal(err)
	}
}

func callAnalyticsReconcile(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	authorized bool,
) (int, int, error) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(),
			"RESET ROLE; RESET SESSION AUTHORIZATION")
		_ = connection.Close()
	}()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	configureAnalyticsReconcileTransaction(t, ctx, transaction, authorized)
	var reconciled, remaining int
	err = transaction.QueryRowContext(ctx, `
SELECT reconciled_count,remaining_count
FROM platform.analytics_reconcile_current_tenant_authorities_v2()`,
	).Scan(&reconciled, &remaining)
	if err != nil {
		return 0, 0, err
	}
	if _, err = transaction.ExecContext(ctx,
		"RESET ROLE; RESET SESSION AUTHORIZATION"); err != nil {
		return 0, 0, err
	}
	if err = transaction.Commit(); err != nil {
		return 0, 0, err
	}
	return reconciled, remaining, nil
}

func reconcileAnalyticsAuthorities(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	authorized bool,
) (int, int) {
	t.Helper()
	reconciled, remaining, err := callAnalyticsReconcile(
		t, ctx, database, authorized,
	)
	if !authorized {
		requireAnalyticsPGCode(t, err, "42501", "authority reconcile capability")
		return 0, 0
	}
	if err != nil {
		t.Fatalf("reconcile Analytics tenant authorities: %v", err)
	}
	return reconciled, remaining
}
