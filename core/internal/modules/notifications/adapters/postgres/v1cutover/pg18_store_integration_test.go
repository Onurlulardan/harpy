//go:build integration

package v1cutover

import (
	"context"
	"testing"
	"time"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func cutoverTestStore(transaction pgx.Tx, tenantID uuid.UUID) *Store {
	return &Store{tenantID: tenantID,
		queryRow: func(ctx context.Context, query string, arguments ...any) platformpostgres.Row {
			return transaction.QueryRow(ctx, query, arguments...)
		},
		query: func(ctx context.Context, query string, arguments ...any) (platformpostgres.Rows, error) {
			return transaction.Query(ctx, query, arguments...)
		},
		execute: func(ctx context.Context, query string, arguments ...any) error {
			_, err := transaction.Exec(ctx, query, arguments...)
			return err
		}, allowed: func(context.Context) bool { return true }}
}

func seedVerifiedRun(t *testing.T, ctx context.Context, transaction pgx.Tx,
	runID uuid.UUID,
) {
	t.Helper()
	_, err := transaction.Exec(ctx, `INSERT INTO notifications_adoption_runs
(id,source_generation,source_manifest_digest,lifecycle_state,started_at,completed_at,row_version)
VALUES ($1,'v1-main',decode(repeat('11',32),'hex'),'verified',$2,$2,2)`,
		runID, time.Now().UTC().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
}
