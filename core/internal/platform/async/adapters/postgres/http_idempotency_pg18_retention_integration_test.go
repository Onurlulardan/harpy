//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/idempotency"
	"github.com/google/uuid"
)

func assertHTTPIdempotencyAsyncRetention(
	t *testing.T,
	ctx context.Context,
	fixture httpIdempotencyPGFixture,
) {
	t.Helper()
	jobID := uuid.New()
	tenantID := uuid.MustParse(asyncTenantA)
	result, err := fixture.runtime.Exec(
		ctx, enqueueJobSQL,
		jobID, tenantID, "tenant", tenantID,
		"operations", int64(1), "workflow", int64(1), "v2",
		"workflow.http_handle", 1, "artifact://payload/1", strings.Repeat("d", 64),
		"http-handle-"+jobID.String(), "database_transactional", nil,
		0, time.Now().UTC(), 8, nil,
	)
	if err != nil || result.RowsAffected() != 1 {
		t.Fatalf("seed async handle job affected %d: %v", result.RowsAffected(), err)
	}
	requested := time.Now().UTC().Add(45 * 24 * time.Hour).Truncate(time.Second)
	request, err := idempotency.NewAsyncHandleRequest(
		"async-key", strings.Repeat("e", 64), 14, jobID, requested,
	)
	if err != nil {
		t.Fatal(err)
	}
	transaction, store := fixture.begin(t, ctx, uuid.MustParse(asyncActor))
	defer transaction.Rollback(ctx)
	decision, err := store.BeginAuthorizedHTTPIdempotency(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	permit, ok := decision.Permit()
	if !ok {
		t.Fatal("async handle request did not return permit")
	}
	response, err := idempotency.NewReferenceResponse(
		202, map[string]string{"retry-after": "1"}, "jobs/"+jobID.String(),
	)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := store.CompleteHTTPIdempotency(ctx, permit, response)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.AsyncJobID() == nil || *receipt.AsyncJobID() != jobID ||
		receipt.RetainUntil().Before(requested) ||
		receipt.RetainUntil().Before(receipt.CompletedAt().Add(30*24*time.Hour)) {
		t.Fatalf("async receipt retention = %#v", receipt)
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
