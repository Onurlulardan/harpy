package v1adoption

import (
	"context"
	"fmt"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
)

func (storage *store) Apply(
	ctx context.Context,
	operation billingv1adoptionapp.Operation,
) (int64, error) {
	var count int64
	if err := storage.database.QueryRow(
		ctx, applyScopeSQL, operation.OperationID, operation.RunID,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("apply Billing V1 adoption scope: %w", err)
	}
	return count, nil
}
