//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/idempotency"
	"github.com/google/uuid"
)

func assertHTTPIdempotencyRollback(
	t *testing.T,
	ctx context.Context,
	fixture httpIdempotencyPGFixture,
) {
	t.Helper()
	transaction, store := fixture.begin(t, ctx, uuid.MustParse(asyncActor))
	defer transaction.Rollback(ctx)
	request, _ := idempotency.NewWriteRequest(
		"rollback-key", strings.Repeat("c", 64), 13, idempotency.AdministrativeWrite,
	)
	decision, err := store.BeginAuthorizedHTTPIdempotency(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	permit, _ := decision.Permit()
	probeID := uuid.New()
	if _, err := transaction.Exec(ctx,
		"INSERT INTO platform.http_idempotency_business_probe(id,value) VALUES($1,'pending')",
		probeID,
	); err != nil {
		t.Fatal(err)
	}
	response, err := idempotency.NewBodyResponse(204, nil, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CompleteHTTPIdempotency(ctx, permit, response); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var receiptCount, probeCount int
	if err := fixture.owner.QueryRowContext(ctx,
		"SELECT count(*) FROM platform.idempotency_receipts WHERE idempotency_key='rollback-key'",
	).Scan(&receiptCount); err != nil {
		t.Fatal(err)
	}
	if err := fixture.owner.QueryRowContext(ctx,
		"SELECT count(*) FROM platform.http_idempotency_business_probe WHERE id=$1", probeID,
	).Scan(&probeCount); err != nil {
		t.Fatal(err)
	}
	if receiptCount != 0 || probeCount != 0 {
		t.Fatalf("rollback left receipt=%d business=%d", receiptCount, probeCount)
	}
}

func assertHTTPIdempotencyPrincipalScope(
	t *testing.T,
	ctx context.Context,
	fixture httpIdempotencyPGFixture,
) {
	t.Helper()
	other := uuid.New()
	transaction, store := fixture.begin(t, ctx, other)
	defer transaction.Rollback(ctx)
	request, _ := idempotency.NewWriteRequest(
		"race-key", strings.Repeat("b", 64), 1, idempotency.InteractiveWrite,
	)
	decision, err := store.BeginAuthorizedHTTPIdempotency(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := decision.Permit(); !ok {
		t.Fatal("different principal shared an idempotency identity")
	}
}
