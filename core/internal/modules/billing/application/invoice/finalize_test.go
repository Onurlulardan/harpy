package invoice

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	invoicedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/invoice"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

type testStore struct {
	snapshot DraftSnapshot
	existing Finalization
	found    bool
	written  []Finalization
	events   []FinalizedEventIntent
}

func (store *testStore) Finalize(
	_ context.Context,
	_ FinalizeExecution,
	planner FinalizePlanner,
) (FinalizeResult, error) {
	var existing *Finalization
	if store.found {
		existing = &store.existing
	}
	plan, err := planner(store.snapshot, existing)
	if err != nil {
		return FinalizeResult{}, err
	}
	if plan.Write != nil {
		store.written = append(store.written, *plan.Write)
	}
	if plan.Event != nil {
		store.events = append(store.events, *plan.Event)
	}
	return plan.Result, nil
}

func TestFinalizeCalculatesAndWritesImmutableTotals(t *testing.T) {
	finalizer, store, request, now := finalizerFixture(t)
	result, err := finalizer.Finalize(context.Background(), request)
	if err != nil || result.Replay || len(store.written) != 1 || len(store.events) != 1 ||
		result.Finalization.Totals.InvoiceTotal().MinorUnits() != 1080 ||
		result.Finalization.Totals.AmountDue().MinorUnits() != 580 ||
		!result.Finalization.FinalizedAt.Equal(now) {
		t.Fatalf("result=%+v written=%+v err=%v", result, store.written, err)
	}
}

func TestFinalizeExactOperationReplays(t *testing.T) {
	finalizer, store, request, now := finalizerFixture(t)
	totals, _ := invoicedomain.CalculateTotals(store.snapshot.Lines, store.snapshot.Settled)
	store.found = true
	store.existing = Finalization{
		InvoiceID: request.InvoiceID, SourceOperationID: request.SourceOperationID,
		ExpectedRevision: request.ExpectedRevision,
		ResultRevision:   request.ExpectedRevision + 1,
		Totals:           totals, PolicySHA256: store.snapshot.PolicySHA256, FinalizedAt: now,
	}
	result, err := finalizer.Finalize(context.Background(), request)
	if err != nil || !result.Replay || len(store.written) != 0 || len(store.events) != 0 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestFinalizeRejectsOperationReuseAndStaleRevision(t *testing.T) {
	finalizer, store, request, now := finalizerFixture(t)
	totals, _ := invoicedomain.CalculateTotals(store.snapshot.Lines, store.snapshot.Settled)
	store.found = true
	store.existing = Finalization{
		InvoiceID: request.InvoiceID, SourceOperationID: request.SourceOperationID,
		ExpectedRevision: request.ExpectedRevision,
		ResultRevision:   request.ExpectedRevision + 1, Totals: totals,
		PolicySHA256: sha256.Sum256([]byte("different")), FinalizedAt: now,
	}
	if _, err := finalizer.Finalize(context.Background(), request); !errors.Is(
		err, ErrOperationReused,
	) {
		t.Fatalf("expected ErrOperationReused, got %v", err)
	}
	store.found = false
	store.snapshot.Revision++
	if _, err := finalizer.Finalize(context.Background(), request); err != ErrState {
		t.Fatalf("expected ErrState, got %v", err)
	}
}

func finalizerFixture(
	t *testing.T,
) (*Finalizer, *testStore, FinalizeRequest, time.Time) {
	t.Helper()
	now := time.Date(2026, time.July, 17, 17, 0, 0, 0, time.UTC)
	zero := amount(t, 0)
	line, err := invoicedomain.NewLine(
		lineID(1), amount(t, 1000), amount(t, 100), amount(t, 180), zero, zero)
	if err != nil {
		t.Fatal(err)
	}
	invoiceID := mustV7(t)
	store := &testStore{snapshot: DraftSnapshot{
		InvoiceID: invoiceID, Revision: 3, Lines: []invoicedomain.Line{line},
		Settled: amount(t, 500), PolicySHA256: sha256.Sum256([]byte("policy")),
	}}
	finalizer, err := NewFinalizer(store, uuid.NewV7, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	request := FinalizeRequest{
		InvoiceID: invoiceID, ActorID: mustV7(t),
		ExpectedRevision: 3, SourceOperationID: mustV7(t),
	}
	return finalizer, store, request, now
}

func amount(t *testing.T, units int64) money.Money {
	t.Helper()
	value, err := money.New(units, "USD", money.RegistryVersion)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func lineID(last byte) invoicedomain.LineIdentifier {
	var value invoicedomain.LineIdentifier
	value[len(value)-1] = last
	return value
}

func mustV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
