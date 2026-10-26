package billingv1adoptionapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Operation struct {
	OperationID uuid.UUID
	RunID       uuid.UUID
	AuthorityID uuid.UUID
	PlanSHA256  string
}

type Reconciliation struct {
	SourceCount      int64
	TargetCount      int64
	SourceSHA256     string
	TargetSHA256     string
	MoneyDifference  int64
	PeriodDifference int64
	CouponDifference int64
	UsageDifference  int64
}

type RollbackRehearsal struct {
	ReceiptID     uuid.UUID
	ReverseSHA256 string
	SourceSHA256  string
	RehearsedAt   time.Time
}

type ScopeStore interface {
	Register(context.Context, Operation, ScopePlan, []byte, CutoverEvidence) (string, error)
	Apply(context.Context, Operation) (int64, error)
	Verify(context.Context, Operation) (Reconciliation, error)
	RehearseRollback(context.Context, Operation, uuid.UUID, string, time.Time) (RollbackRehearsal, error)
}

func validOperation(value Operation) bool {
	return value.OperationID != uuid.Nil && value.RunID != uuid.Nil &&
		value.AuthorityID != uuid.Nil && validDigest(value.PlanSHA256)
}
