//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func assertAuthorityCapabilityFlow(
	t *testing.T, ctx context.Context, database *sql.DB,
	plan transactionCapabilityPlan,
) {
	t.Helper()
	if err := callAuthorityCapabilityWrapper(
		ctx, database, plan.operationID, strings.Repeat("f", 64),
	); err == nil {
		t.Fatal("authority wrapper accepted the wrong manifest")
	}
	var wait sync.WaitGroup
	errorsFound := make(chan error, 2)
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			errorsFound <- callAuthorityCapabilityWrapper(
				ctx, database, plan.operationID, plan.manifest,
			)
		}()
	}
	wait.Wait()
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatalf("concurrent authority wrapper: %v", err)
		}
	}
	var calls, scopes int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT pg_catalog.count(*) FROM platform.test_tenancy_authority_calls
 WHERE operation_id=$1),
(SELECT pg_catalog.count(*) FROM platform.tenancy_authority_reconcile_scopes
 WHERE operation_id=$1)`, plan.operationID).Scan(&calls, &scopes)
	if err != nil || calls != 3 || scopes != 0 {
		t.Fatalf("authority calls/scopes=%d/%d err=%v", calls, scopes, err)
	}
}

func callAuthorityCapabilityWrapper(
	ctx context.Context, database *sql.DB, operationID uuid.UUID, manifest string,
) error {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime;
SET LOCAL ROLE harpy_runtime;
SELECT pg_catalog.set_config(
 'application_name','harpy-v2:core-operator-initialize',true
)`); err != nil {
		return err
	}
	rows := make([]int, 6)
	err = transaction.QueryRowContext(ctx, `
SELECT * FROM platform.tenancy_authority_reconcile_current_v1($1,$2)`,
		operationID, manifest).Scan(
		&rows[0], &rows[1], &rows[2], &rows[3],
		&rows[4], &rows[5],
	)
	if err != nil {
		return err
	}
	return transaction.Commit()
}
