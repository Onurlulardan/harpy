package v1adoption

import (
	"context"
	"fmt"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
)

func (storage *store) Verify(
	ctx context.Context,
	operation billingv1adoptionapp.Operation,
) (billingv1adoptionapp.Reconciliation, error) {
	var result billingv1adoptionapp.Reconciliation
	err := storage.database.QueryRow(
		ctx, verifyScopeSQL, operation.OperationID, operation.RunID,
	).Scan(
		&result.SourceCount, &result.TargetCount,
		&result.SourceSHA256, &result.TargetSHA256,
		&result.MoneyDifference, &result.PeriodDifference,
		&result.CouponDifference, &result.UsageDifference,
	)
	if err != nil {
		return billingv1adoptionapp.Reconciliation{},
			fmt.Errorf("verify Billing V1 adoption scope: %w", err)
	}
	return result, nil
}
