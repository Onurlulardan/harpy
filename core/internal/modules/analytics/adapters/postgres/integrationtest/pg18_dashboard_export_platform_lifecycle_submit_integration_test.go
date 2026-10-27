//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

func submitAnalyticsDashboardExportFixture(t *testing.T, ctx context.Context,
	connection *sql.Conn, operationID, exportID, jobID, expiryJobID,
	idempotencyKey string, marker byte, occurredAt, expiresAt time.Time,
) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var storedID, state string
	var replayed bool
	err := transaction.QueryRowContext(ctx, `SELECT export_id,effective_state,replayed
FROM analytics_dashboard_export_submit_v2('t_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8,
'csv','{}'::bytea,1,1,1,100000,67108864,$9,$10)`, operationID,
		analyticsDashboardActor, idempotencyKey, bytes.Repeat([]byte{marker}, 32),
		exportID, jobID, expiryJobID, analyticsDashboard, occurredAt, expiresAt,
	).Scan(&storedID, &state, &replayed)
	if err != nil || storedID != exportID || state != "queued" || replayed {
		t.Fatalf("submit Platform-first export %s = %s/%s/%v: %v",
			exportID, storedID, state, replayed, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
