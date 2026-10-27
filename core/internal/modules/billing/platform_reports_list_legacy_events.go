package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	billingreports "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/reports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (reports *platformReports) ListLegacyEvents(
	ctx context.Context,
	query billingapi.Query,
) (billingapi.Page[billingapi.LegacyEventView], error) {
	var result billingreports.RecordResult
	err := reports.inSnapshot(
		ctx, query.Access, billingapi.PermissionPlatformBillingAudit,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			var queryErr error
			result, queryErr = reports.queries.ExecuteLegacyEvents(
				scoped, scope, query.Request,
			)
			return queryErr
		},
	)
	if err != nil {
		return billingapi.Page[billingapi.LegacyEventView]{}, normalizePlatformReportError(err)
	}
	items := result.Items()
	views := make([]billingapi.LegacyEventView, len(items))
	for index, item := range items {
		views[index] = billingapi.NewLegacyEventQueryView(item)
	}
	return recordPage(views, result, query), nil
}
