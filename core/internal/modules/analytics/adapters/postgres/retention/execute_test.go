package retention

import (
	"context"
	"testing"
	"time"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/retention"
	"github.com/google/uuid"
)

type retentionRowStub struct{ values []any }

func (row retentionRowStub) Scan(destinations ...any) error {
	for index, value := range row.values {
		switch destination := destinations[index].(type) {
		case *uuid.UUID:
			*destination = value.(uuid.UUID)
		case *time.Time:
			*destination = value.(time.Time)
		case *string:
			*destination = value.(string)
		case *int32:
			*destination = value.(int32)
		case *int64:
			*destination = value.(int64)
		case *bool:
			*destination = value.(bool)
		}
	}
	return nil
}

func TestExecuteReturnsPersistedBoundedOutcome(t *testing.T) {
	tenantID := uuid.MustParse("019b1000-0000-7000-8000-000000000001")
	jobID := uuid.MustParse("019b1000-0000-7000-8000-000000000101")
	attemptID := uuid.MustParse("019b1000-0000-7000-8000-000000000102")
	nextID := uuid.MustParse("019b1000-0000-7000-8000-000000000103")
	reference, digest, _ := retentionapp.Reference(tenantID)
	storage, err := NewStore(DatabaseAccess{QueryRow: func(
		_ context.Context, _ string, _ ...any,
	) Row {
		return retentionRowStub{values: []any{
			nextID, time.Date(2026, 7, 20, 3, 0, 0, 0, time.UTC),
			reference, digest, "analytics:retention:" + nextID.String(), int32(0),
			int64(1), int64(2), int64(3), int64(4), int64(5), int64(6),
			false, true, int64(7),
		}}
	}})
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := storage.Execute(context.Background(), retentionapp.Claim{
		JobID: jobID, AttemptID: attemptID, TenantID: tenantID,
		Kind: retentionapp.JobKind, PayloadRevision: 1,
		PayloadReference: reference, PayloadSHA256: digest,
	}, nextID)
	if err != nil || outcome.CheckpointRevision != 7 ||
		outcome.Counts.Capabilities != 6 || outcome.Successor.JobID != nextID {
		t.Fatalf("retention adapter outcome=%+v err=%v", outcome, err)
	}
}
