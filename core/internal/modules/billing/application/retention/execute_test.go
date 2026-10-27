package retention

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type storeStub struct {
	outcome Outcome
	err     error
	claim   Claim
	nextID  uuid.UUID
}

func (stub *storeStub) Execute(
	_ context.Context, claim Claim, nextID uuid.UUID,
) (Outcome, error) {
	stub.claim, stub.nextID = claim, nextID
	return stub.outcome, stub.err
}

func TestExecuteValidatesClaimAndPersistedSuccessor(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	jobID := uuid.MustParse("018f1000-0000-7000-8000-000000000001")
	attemptID := uuid.MustParse("018f1000-0000-7000-8000-000000000002")
	nextID := uuid.MustParse("018f1000-0000-7000-8000-000000000003")
	reference, digest, err := Reference(tenantID)
	if err != nil {
		t.Fatal(err)
	}
	claim := Claim{
		JobID: jobID, AttemptID: attemptID, TenantID: tenantID, Kind: JobKind,
		PayloadRevision: PayloadRevision, PayloadReference: reference,
		PayloadSHA256: digest,
	}
	stub := &storeStub{outcome: Outcome{
		Successor: Successor{
			JobID: nextID, RunAt: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
			PayloadReference: reference, PayloadSHA256: digest,
			DedupeKey: "billing:retention:" + nextID.String(),
		},
		RawBodiesRedacted: 2, RawRecordsPurged: 3, CheckpointRevision: 4,
	}}
	commands, err := NewCommands(stub, func() (uuid.UUID, error) { return nextID, nil })
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := commands.Execute(context.Background(), claim)
	if err != nil || outcome.CheckpointRevision != 4 || stub.claim != claim ||
		stub.nextID != nextID {
		t.Fatalf("retention execute outcome=%+v err=%v", outcome, err)
	}
}

func TestExecuteRejectsInvalidClaimAndOutcome(t *testing.T) {
	commands, err := NewCommands(&storeStub{}, uuid.NewV7)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := commands.Execute(context.Background(), Claim{}); !errors.Is(err, ErrClaimInvalid) {
		t.Fatalf("invalid claim error=%v", err)
	}
}
