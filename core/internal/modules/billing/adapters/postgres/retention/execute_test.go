package retention

import (
	"context"
	"testing"
	"time"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/retention"
	"github.com/google/uuid"
)

type rowStub struct {
	values []any
	err    error
}

func (row rowStub) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
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
		}
	}
	return nil
}

func TestExecuteUsesCurrentTenantAuthorityAndReturnsCycleOutcome(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	authorityID := uuid.MustParse("018f1000-0000-7000-8000-000000000010")
	jobID := uuid.MustParse("018f1000-0000-7000-8000-000000000011")
	attemptID := uuid.MustParse("018f1000-0000-7000-8000-000000000012")
	nextID := uuid.MustParse("018f1000-0000-7000-8000-000000000013")
	reference, digest, _ := retentionapp.Reference(tenantID)
	queries := 0
	storage, err := NewStore(DatabaseAccess{QueryRow: func(
		_ context.Context, statement string, _ ...any,
	) Row {
		queries++
		if statement == resolveAuthoritySQL {
			return rowStub{values: []any{authorityID}}
		}
		return rowStub{values: []any{
			nextID, time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
			reference, digest, "billing:retention:" + nextID.String(), int32(0),
			int64(1), int64(2), int64(3), int64(4), int64(5), int64(6),
		}}
	}})
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := storage.Execute(context.Background(), retentionapp.Claim{
		JobID: jobID, AttemptID: attemptID, TenantID: tenantID,
	}, nextID)
	if err != nil || queries != 2 || outcome.CheckpointRevision != 6 ||
		outcome.Successor.JobID != nextID {
		t.Fatalf("retention adapter outcome=%+v queries=%d err=%v", outcome, queries, err)
	}
}
