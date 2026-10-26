package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	billingreports "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/reports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (reports *platformReports) ListAuditEvents(
	ctx context.Context,
	query billingapi.Query,
) (billingapi.Page[billingapi.AuditEventView], error) {
	var result billingreports.RecordResult
	err := reports.inSnapshot(
		ctx, query.Access, billingapi.PermissionPlatformBillingAudit,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			var queryErr error
			result, queryErr = reports.queries.ReadAuditEvents(scoped, scope, query.Request)
			return queryErr
		},
	)
	if err != nil {
		return billingapi.Page[billingapi.AuditEventView]{}, normalizePlatformReportError(err)
	}
	items := result.Items()
	views := make([]billingapi.AuditEventView, len(items))
	for index, item := range items {
		views[index] = billingapi.NewAuditEventQueryView(item)
	}
	return recordPage(views, result, query), nil
}
