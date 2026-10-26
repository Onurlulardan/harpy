//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"
)

func assertAsyncWriterSQL(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	runAt := time.Now().UTC().Truncate(time.Microsecond)
	jobArguments := []any{
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", asyncTenantA, "tenant", asyncTenantA,
		"operations", int64(1), "workflow", int64(1), "v2", "test.job", 1,
		"payload/jobs/writer", asyncSHA256([]byte("job payload")), "writer-sql-job",
		"database_transactional", nil, 5, runAt, 3, nil,
	}
	assertAsyncInsertThenConflict(t, ctx, database, enqueueJobSQL, jobArguments)
	var jobID string
	var jobExact bool
	err := database.QueryRowContext(
		ctx, findJobEnqueueSQL,
		jobArguments[0], asyncTenantA, "operations", "workflow", "test.job",
		"writer-sql-job", int64(1), int64(1), 1, asyncSHA256([]byte("job payload")),
		"database_transactional", nil, 5, 3, nil, runAt, "payload/jobs/writer",
	).Scan(&jobID, &jobExact)
	if err != nil || !jobExact || jobID != jobArguments[0] {
		t.Fatalf("find exact job enqueue = %s, %t, %v", jobID, jobExact, err)
	}

	payload := []byte(`{"ok":true}`)
	digest := sha256.Sum256(payload)
	payloadSHA := hex.EncodeToString(digest[:])
	occurredAt := runAt.Add(-time.Second)
	eventArguments := []any{
		"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", asyncTenantA, "tenant", asyncTenantA,
		"operations", int64(1), "workflow", int64(1), "v2", "integration-event.v1",
		"test.event", 1, "test.aggregate", "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
		int64(1), int64(1), occurredAt, asyncActor, "writer-sql",
		asyncCausation, "writer-event", "writer-effect", payload, payloadSHA,
		asyncSHA256([]byte("event schema")),
	}
	assertAsyncInsertThenConflict(t, ctx, database, enqueueEventSQL, eventArguments)
	var eventID string
	var eventExact bool
	err = database.QueryRowContext(
		ctx, findEventEnqueueSQL,
		eventArguments[0], asyncTenantA, "test.aggregate", eventArguments[13], int64(1),
		"operations", int64(1), "workflow", int64(1), "integration-event.v1",
		"test.event", 1, int64(1), asyncActor, "writer-sql", asyncCausation,
		"writer-event", "writer-effect", payloadSHA,
		asyncSHA256([]byte("event schema")), occurredAt,
	).Scan(&eventID, &eventExact)
	if err != nil || !eventExact || eventID != eventArguments[0] {
		t.Fatalf("find exact event enqueue = %s, %t, %v", eventID, eventExact, err)
	}
}

func assertAsyncInsertThenConflict(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	statement string,
	arguments []any,
) {
	t.Helper()
	for attempt, wanted := range []int64{1, 0} {
		result, err := database.ExecContext(ctx, statement, arguments...)
		if err != nil {
			t.Fatalf("writer SQL attempt %d: %v", attempt+1, err)
		}
		rows, err := result.RowsAffected()
		if err != nil || rows != wanted {
			t.Fatalf("writer SQL rows = %d, want %d, error %v", rows, wanted, err)
		}
	}
}
