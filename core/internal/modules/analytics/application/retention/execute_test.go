package retention

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type retentionStoreStub struct {
	outcome Outcome
	claim   Claim
	nextID  uuid.UUID
}

func (stub *retentionStoreStub) Execute(
	_ context.Context, claim Claim, nextID uuid.UUID,
) (Outcome, error) {
	stub.claim, stub.nextID = claim, nextID
	return stub.outcome, nil
}

func TestExecutePinsClaimAndPersistedSuccessor(t *testing.T) {
	tenantID := uuid.MustParse("019b1000-0000-7000-8000-000000000001")
	jobID := uuid.MustParse("019b1000-0000-7000-8000-000000000101")
	attemptID := uuid.MustParse("019b1000-0000-7000-8000-000000000102")
	nextID := uuid.MustParse("019b1000-0000-7000-8000-000000000103")
	reference, digest, _ := Reference(tenantID)
	claim := Claim{JobID: jobID, AttemptID: attemptID, TenantID: tenantID,
		Kind: JobKind, PayloadRevision: 1, PayloadReference: reference,
		PayloadSHA256: digest}
	stub := &retentionStoreStub{outcome: Outcome{
		Counts: Counts{Drafts: 2, Versions: 1}, CheckpointRevision: 4,
		Successor: Successor{JobID: nextID,
			RunAt:            time.Date(2026, 7, 19, 3, 0, 0, 0, time.UTC),
			PayloadReference: reference, PayloadSHA256: digest,
			DedupeKey: "analytics:retention:" + nextID.String()},
	}}
	commands, err := NewCommands(stub, func() (uuid.UUID, error) { return nextID, nil })
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := commands.Execute(context.Background(), claim)
	if err != nil || outcome.CheckpointRevision != 4 || stub.claim != claim ||
		stub.nextID != nextID {
		t.Fatalf("retention outcome=%+v claim=%+v err=%v", outcome, stub.claim, err)
	}
}

func TestExecuteRejectsInvalidClaimAndUnboundedOutcome(t *testing.T) {
	commands, err := NewCommands(&retentionStoreStub{}, uuid.NewV7)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = commands.Execute(context.Background(), Claim{}); !errors.Is(err, ErrClaimInvalid) {
		t.Fatalf("invalid claim error=%v", err)
	}
	tenantID := uuid.MustParse("019b1000-0000-7000-8000-000000000001")
	reference, digest, _ := Reference(tenantID)
	outcome := Outcome{Counts: Counts{Drafts: 101}, CheckpointRevision: 1,
		Successor: Successor{JobID: uuid.MustParse("019b1000-0000-7000-8000-000000000103"),
			RunAt:            time.Now().UTC().Truncate(time.Microsecond),
			PayloadReference: reference, PayloadSHA256: digest, DedupeKey: "next"}}
	if !errors.Is(outcome.Validate(tenantID), ErrOutcomeInvalid) {
		t.Fatal("unbounded retention outcome was accepted")
	}
}
