package profilebackfill

import (
	"context"
	"errors"
	"time"
)

const (
	RecoveryInventoryReadBatchSize         = 64
	MaximumRecoveryInventoryRows     int64 = 100_000
	MaximumRecoveryInventoryBytes    int64 = 256 << 20
	MaximumRecoveryInventoryDuration       = 30 * time.Second
)

var ErrBudgetExceeded = errors.New("IAM profile recovery inventory budget exceeded")

type RecoveryInventoryClock func() time.Time

type RecoveryInventoryBudget struct {
	clock     RecoveryInventoryClock
	startedAt time.Time
	expiresAt time.Time
	rows      int64
	bytes     int64
}

func NewRecoveryInventoryBudget(
	clock RecoveryInventoryClock,
) (*RecoveryInventoryBudget, error) {
	if clock == nil {
		return nil, ErrUnavailable
	}
	startedAt := clock().UTC()
	if startedAt.IsZero() {
		return nil, ErrUnavailable
	}
	return &RecoveryInventoryBudget{
		clock: clock, startedAt: startedAt,
		expiresAt: startedAt.Add(MaximumRecoveryInventoryDuration),
	}, nil
}

func (budget *RecoveryInventoryBudget) Check(ctx context.Context) error {
	if budget == nil || budget.clock == nil || ctx == nil {
		return ErrInvalid
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(context.Cause(ctx), ErrBudgetExceeded) {
			return ErrBudgetExceeded
		}
		return err
	}
	now := budget.clock().UTC()
	if now.IsZero() || now.Before(budget.startedAt) {
		return ErrUnavailable
	}
	if !now.Before(budget.expiresAt) {
		return ErrBudgetExceeded
	}
	return nil
}

func (budget *RecoveryInventoryBudget) Admit(
	ctx context.Context,
	documentBytes int,
) error {
	if err := budget.Check(ctx); err != nil {
		return err
	}
	if documentBytes < 2 {
		return ErrInvalid
	}
	if budget.rows >= MaximumRecoveryInventoryRows ||
		int64(documentBytes) > MaximumRecoveryInventoryBytes-budget.bytes {
		return ErrBudgetExceeded
	}
	budget.rows++
	budget.bytes += int64(documentBytes)
	return nil
}
