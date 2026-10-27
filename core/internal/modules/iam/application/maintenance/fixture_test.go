package maintenance

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	fixtureTenantID = uuid.MustParse("019c6a7e-7c00-7000-8000-000000000001")
	fixtureJobID    = uuid.MustParse("019c6a7e-7c00-7000-8000-000000000002")
	fixtureAttempt  = uuid.MustParse("019c6a7e-7c00-7000-8000-000000000003")
	fixtureNextID   = uuid.MustParse("019c6a7e-7c00-7000-8000-000000000004")
	fixturePlanID   = uuid.MustParse("019c6a7e-7c00-7000-8000-000000000005")
)

type fixtureStore struct{ called Kind }

func (store *fixtureStore) result(claim Claim, nextID uuid.UUID) (Outcome, error) {
	store.called = claim.Kind
	if !claim.Kind.Recurring() {
		return Outcome{}, nil
	}
	return Outcome{Successor: &Successor{
		JobID: nextID, RunAt: time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC),
		PayloadReference: claim.PayloadReference,
		PayloadSHA256:    claim.PayloadSHA256, DedupeKey: "iam:maintenance:next",
	}}, nil
}

func fixtureClaim(t *testing.T, kind Kind) Claim {
	t.Helper()
	var reference, digest string
	var err error
	if kind.Recurring() {
		reference, digest, err = RecurringReference(fixtureTenantID, kind)
	} else {
		reference, digest, err = ProfileReference(fixtureTenantID, fixturePlanID, 1)
	}
	if err != nil {
		t.Fatalf("build reference: %v", err)
	}
	return Claim{JobID: fixtureJobID, AttemptID: fixtureAttempt,
		TenantID: fixtureTenantID, Kind: kind, PayloadRevision: PayloadRevision,
		PayloadReference: reference, PayloadSHA256: digest}
}

func executeFixture(t *testing.T, kind Kind) {
	t.Helper()
	store := &fixtureStore{}
	commands, err := NewCommands(store, func() (uuid.UUID, error) {
		return fixtureNextID, nil
	})
	if err != nil {
		t.Fatalf("new commands: %v", err)
	}
	outcome, err := commands.Execute(context.Background(), fixtureClaim(t, kind))
	if err != nil || store.called != kind || outcome.Validate(kind) != nil {
		t.Fatalf("execute %s = %+v, called %s, %v", kind, outcome, store.called, err)
	}
}

func (store *fixtureStore) Execute(
	ctx context.Context, claim Claim, id uuid.UUID,
) (Outcome, error) {
	return store.result(claim, id)
}
