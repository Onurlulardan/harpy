package retention

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type storeFake struct {
	cycle Cycle
}

func (store *storeFake) ApplyCycle(_ context.Context, cycle Cycle) (Outcome, error) {
	store.cycle = cycle
	reference, digest, _ := Reference(cycle.Claim.TenantID)
	return Outcome{Counts: Counts{ContentRedacted: 2}, CheckpointRevision: 3,
		Successor: Successor{JobID: cycle.SuccessorJobID,
			RunAt: cycle.CompletedAt.Add(time.Minute), PayloadReference: reference,
			PayloadSHA256: digest, DedupeKey: "notifications:retention:next"}}, nil
}

func TestPurgeValidatesClaimAndPersistsCycle(t *testing.T) {
	tenant := uuid.MustParse("019b0000-0000-7000-8000-000000000061")
	reference, digest, _ := Reference(tenant)
	claim := Claim{
		JobID:     uuid.MustParse("019b0000-0000-7000-8000-000000000062"),
		AttemptID: uuid.MustParse("019b0000-0000-7000-8000-000000000063"),
		TenantID:  tenant, Kind: JobKind, PayloadRevision: PayloadRevision,
		PayloadReference: reference, PayloadSHA256: digest,
	}
	now := time.Date(2028, 1, 2, 4, 0, 0, 123456000, time.UTC)
	next := uuid.MustParse("019b0000-0000-7000-8000-000000000064")
	store := &storeFake{}
	commands, err := NewCommands(store, func() (uuid.UUID, error) { return next, nil },
		func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := commands.Purge(t.Context(), claim)
	if err != nil || outcome.Counts.ContentRedacted != 2 ||
		store.cycle.Claim.JobID != claim.JobID || store.cycle.CompletedAt != now {
		t.Fatalf("outcome=%+v cycle=%+v err=%v", outcome, store.cycle, err)
	}
	claim.PayloadSHA256 = "invalid"
	if _, err = commands.Purge(t.Context(), claim); err != ErrClaimInvalid {
		t.Fatalf("invalid claim error=%v", err)
	}
}
