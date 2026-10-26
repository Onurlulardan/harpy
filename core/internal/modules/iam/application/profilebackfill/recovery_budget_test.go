package profilebackfill

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRecoveryInventoryBudgetEnforcesRowsAndBytes(t *testing.T) {
	now := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	rowBudget, err := NewRecoveryInventoryBudget(func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	for index := int64(0); index < MaximumRecoveryInventoryRows; index++ {
		if err := rowBudget.Admit(context.Background(), 2); err != nil {
			t.Fatalf("row %d: %v", index, err)
		}
	}
	if err := rowBudget.Admit(context.Background(), 2); !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("row budget error = %v", err)
	}
	byteBudget, err := NewRecoveryInventoryBudget(func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	const documentBytes = 1 << 20
	for admitted := int64(0); admitted < MaximumRecoveryInventoryBytes; admitted += documentBytes {
		if err := byteBudget.Admit(context.Background(), documentBytes); err != nil {
			t.Fatalf("byte %d: %v", admitted, err)
		}
	}
	if err := byteBudget.Admit(context.Background(), 2); !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("byte budget error = %v", err)
	}
}

func TestRecoveryInventoryBudgetUsesClockAndContextCause(t *testing.T) {
	startedAt := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	now := startedAt
	budget, err := NewRecoveryInventoryBudget(func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	now = startedAt.Add(MaximumRecoveryInventoryDuration)
	if err := budget.Check(context.Background()); !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("time budget error = %v", err)
	}
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(ErrBudgetExceeded)
	if err := budget.Check(ctx); !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("context budget error = %v", err)
	}
}
