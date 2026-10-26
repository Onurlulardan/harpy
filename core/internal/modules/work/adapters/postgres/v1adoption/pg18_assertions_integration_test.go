//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func seedFreshCanonical(t *testing.T, ctx context.Context, database *sql.DB,
	schema, name string, seed byte,
) {
	t.Helper()
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		now := time.Date(2026, 7, 19, 9, 0, 0, 0, time.UTC)
		_, err := transaction.ExecContext(ctx, `
INSERT INTO work_boards
(id,creator_user_id,name,description,lifecycle_state,reference_epoch,row_version,
 created_at,updated_at,archived_at)
VALUES ($1,$2,$3,'fresh board','ACTIVE',1,1,$4,$4,NULL)`,
			workUUID(seed, 1), workUUID(seed, 2), name, now)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func assertSchemaBoard(t *testing.T, ctx context.Context, database *sql.DB,
	schema, expected, forbidden string,
) {
	t.Helper()
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		var expectedCount, forbiddenCount int
		if err := transaction.QueryRowContext(ctx,
			"SELECT count(*) FROM work_boards WHERE name=$1", expected).Scan(&expectedCount); err != nil {
			t.Fatal(err)
		}
		if err := transaction.QueryRowContext(ctx,
			"SELECT count(*) FROM work_boards WHERE name=$1", forbidden).Scan(&forbiddenCount); err != nil {
			t.Fatal(err)
		}
		if expectedCount != 1 || forbiddenCount != 0 {
			t.Fatalf("%s board isolation expected=%d forbidden=%d", schema,
				expectedCount, forbiddenCount)
		}
	})
}

func assertAdoptionRolledBack(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		var boards, runs int
		var fence string
		if err := transaction.QueryRowContext(ctx, "SELECT count(*) FROM work_boards").Scan(&boards); err != nil {
			t.Fatal(err)
		}
		if err := transaction.QueryRowContext(ctx, "SELECT count(*) FROM work_adoption_runs").Scan(&runs); err != nil {
			t.Fatal(err)
		}
		if err := transaction.QueryRowContext(ctx,
			"SELECT state FROM work_v1_writer_fence WHERE singleton").Scan(&fence); err != nil {
			t.Fatal(err)
		}
		if boards != 0 || runs != 0 || fence != "V1_ACTIVE" {
			t.Fatalf("rollback boards=%d runs=%d fence=%s", boards, runs, fence)
		}
	})
}
