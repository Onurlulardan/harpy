package billingv1adoptionapp

import (
	"context"
	"fmt"
)

func ApplyScope(
	ctx context.Context,
	store ScopeStore,
	operation Operation,
) (int64, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || !validOperation(operation) {
		return 0, fmt.Errorf("Billing V1 adoption apply is not configured")
	}
	count, err := store.Apply(ctx, operation)
	if err != nil {
		return 0, fmt.Errorf("apply Billing V1 adoption plan: %w", err)
	}
	if count < 0 {
		return 0, fmt.Errorf("Billing V1 adoption apply count is invalid")
	}
	return count, nil
}
