package requestattribution

import (
	"context"
	"errors"
	"testing"
)

func TestAttributionIsSingleAssignmentAndConflictFailClosed(t *testing.T) {
	ctx, err := WithSlot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	actor, tenant := [16]byte{1}, [16]byte{2}
	if err := Bind(ctx, actor, tenant); err != nil {
		t.Fatal(err)
	}
	if err := Bind(ctx, actor, tenant); err != nil {
		t.Fatalf("idempotent bind failed: %v", err)
	}
	identity, status := Snapshot(ctx)
	if status != StatusVerified || identity.ActorID != actor || identity.TenantID != tenant {
		t.Fatalf("verified attribution = %#v, %d", identity, status)
	}
	if err := Bind(ctx, actor, [16]byte{3}); !errors.Is(err, ErrConflict) {
		t.Fatalf("conflicting bind error = %v", err)
	}
	identity, status = Snapshot(ctx)
	if status != StatusConflict || identity.Valid() {
		t.Fatalf("conflicting attribution leaked: %#v, %d", identity, status)
	}
}

func TestAttributionSurvivesContextDetach(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	ctx, err := WithSlot(parent)
	if err != nil {
		t.Fatal(err)
	}
	if err := Bind(ctx, [16]byte{1}, [16]byte{2}); err != nil {
		t.Fatal(err)
	}
	cancel()
	identity, status := Snapshot(context.WithoutCancel(ctx))
	if status != StatusVerified || !identity.Valid() {
		t.Fatalf("detached attribution = %#v, %d", identity, status)
	}
}

func TestAttributionWithoutSlotIsUnavailable(t *testing.T) {
	if err := Bind(context.Background(), [16]byte{1}, [16]byte{2}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("bind error = %v", err)
	}
	if _, status := Snapshot(context.Background()); status != StatusUnavailable {
		t.Fatalf("status = %d", status)
	}
}
