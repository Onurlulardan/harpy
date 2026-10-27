package billingv1adoptionapp

import (
	"context"
	"fmt"
)

func VerifyScope(
	ctx context.Context,
	store ScopeStore,
	operation Operation,
) (Reconciliation, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || !validOperation(operation) {
		return Reconciliation{}, fmt.Errorf("Billing V1 reconciliation is not configured")
	}
	result, err := store.Verify(ctx, operation)
	if err != nil {
		return Reconciliation{}, fmt.Errorf("verify Billing V1 adoption: %w", err)
	}
	if result.SourceCount != result.TargetCount ||
		result.SourceSHA256 != result.TargetSHA256 ||
		result.MoneyDifference != 0 || result.PeriodDifference != 0 ||
		result.CouponDifference != 0 || result.UsageDifference != 0 {
		return Reconciliation{}, fmt.Errorf("Billing V1 adoption has non-zero reconciliation")
	}
	return result, nil
}
