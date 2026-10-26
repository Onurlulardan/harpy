//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func assertSingleProtectedFinalizationReceipt(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	evidence finalizationEvidence,
) {
	t.Helper()
	var count int
	err := admin.QueryRow(ctx, `SELECT count(*)::integer
FROM cutover_authority.finalization_receipts
WHERE operation_id=$1 AND run_id=$2 AND finalizer_lease_id=$3`,
		evidence.operationID, evidence.runID, evidence.leaseID,
	).Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("protected finalization receipt count=%d err=%v", count, err)
	}
}
