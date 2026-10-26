package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	billingreports "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/reports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (reports *platformReports) GetRevenue(
	ctx context.Context,
	query billingapi.Query,
) (billingapi.RevenueResult, error) {
	var result billingreports.RevenueResult
	err := reports.inSnapshot(
		ctx, query.Access, billingapi.PermissionPlatformBillingRevenue,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			var queryErr error
			result, queryErr = reports.queries.GetRevenue(scoped, scope, query.Request)
			return queryErr
		},
	)
	if err != nil {
		return billingapi.RevenueResult{}, normalizePlatformReportError(err)
	}
	return billingapi.RevenueResult{Groups: result.Groups(), Meta: map[string]any{}}, nil
}
