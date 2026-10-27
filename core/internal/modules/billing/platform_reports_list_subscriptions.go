package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	billingreports "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/reports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (reports *platformReports) ListSubscriptions(
	ctx context.Context,
	query billingapi.Query,
) (billingapi.Page[billingapi.SubscriptionView], error) {
	var result billingreports.RecordResult
	action := func(scoped context.Context, scope platformtx.GlobalScope) error {
		var queryErr error
		result, queryErr = reports.queries.ListSubscriptions(scoped, scope, query.Request)
		return queryErr
	}
	var err error
	if query.Access.ValidFor(billingapi.PermissionPlatformBillingRead) {
		err = reports.inSnapshot(
			ctx, query.Access, billingapi.PermissionPlatformBillingRead, action,
		)
	} else {
		err = reports.inCommandSnapshot(
			ctx, query.Access, billingreports.SubscriptionTransitionOperation, action,
		)
	}
	if err != nil {
		return billingapi.Page[billingapi.SubscriptionView]{}, normalizePlatformReportError(err)
	}
	items := result.Items()
	views := make([]billingapi.SubscriptionView, len(items))
	for index, item := range items {
		views[index] = billingapi.NewSubscriptionQueryView(item)
	}
	return recordPage(views, result, query), nil
}
