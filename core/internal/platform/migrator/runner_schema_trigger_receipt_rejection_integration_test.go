//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertAttemptReceiptRejected(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	jobID string,
	state string,
	code string,
) {
	t.Helper()
	finished := time.Now().UTC().Add(-time.Millisecond).Truncate(time.Microsecond)
	_, err := database.ExecContext(ctx, schemaAttemptInsertSQL,
		jobID, 0, finished, state, integrationDigest(jobID+"-forged-outcome"),
		code, integrationDigest(jobID+"-forged-detail"))
	if err == nil {
		t.Fatalf("forged %s receipt with code %s was accepted", state, code)
	}
}
