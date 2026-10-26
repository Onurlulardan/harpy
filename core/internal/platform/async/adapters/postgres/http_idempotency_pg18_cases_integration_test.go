//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/idempotency"
	"github.com/google/uuid"
)

func assertHTTPIdempotencyRaceAndReplay(
	t *testing.T,
	ctx context.Context,
	fixture httpIdempotencyPGFixture,
) {
	t.Helper()
	principal := uuid.MustParse(asyncActor)
	request, _ := idempotency.NewWriteRequest(
		"race-key", strings.Repeat("a", 64), 11, idempotency.InteractiveWrite,
	)
	firstTX, first := fixture.begin(t, ctx, principal)
	defer firstTX.Rollback(ctx)
	decision, err := first.BeginAuthorizedHTTPIdempotency(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	permit, ok := decision.Permit()
	if !ok {
		t.Fatal("new request did not return a permit")
	}
	secondTX, second := fixture.begin(t, ctx, principal)
	defer secondTX.Rollback(ctx)
	if _, err := second.BeginAuthorizedHTTPIdempotency(
		ctx, request,
	); !errors.Is(err, idempotency.ErrInProgress) {
		t.Fatalf("contending request error = %v", err)
	}
	_ = secondTX.Rollback(ctx)
	response, err := idempotency.NewBodyResponse(
		201, map[string]string{"content-type": "application/json"}, []byte(`{"id":"42"}`),
	)
	if err != nil {
		t.Fatalf("construct response: %v", err)
	}
	if err := response.Validate(); err != nil {
		t.Fatalf("validate response: %v", err)
	}
	receipt, err := first.CompleteHTTPIdempotency(ctx, permit, response)
	if err != nil {
		t.Fatal(err)
	}
	if err := firstTX.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	replayTX, replayStore := fixture.begin(t, ctx, principal)
	defer replayTX.Rollback(ctx)
	replayed, err := replayStore.BeginAuthorizedHTTPIdempotency(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	replay, ok := replayed.Replay()
	if !ok || replay.ID() != receipt.ID() ||
		string(replay.Response().Body()) != `{"id":"42"}` {
		t.Fatalf("replay = %#v", replay)
	}
	_ = replayTX.Rollback(ctx)
	assertHTTPIdempotencyDifferentHash(t, ctx, fixture, principal)
}

func assertHTTPIdempotencyDifferentHash(
	t *testing.T,
	ctx context.Context,
	fixture httpIdempotencyPGFixture,
	principal uuid.UUID,
) {
	t.Helper()
	request, _ := idempotency.NewWriteRequest(
		"race-key", strings.Repeat("b", 64), 12, idempotency.InteractiveWrite,
	)
	transaction, store := fixture.begin(t, ctx, principal)
	defer transaction.Rollback(ctx)
	if _, err := store.BeginAuthorizedHTTPIdempotency(
		ctx, request,
	); !errors.Is(err, idempotency.ErrKeyReused) {
		t.Fatalf("different hash error = %v", err)
	}
}
