package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	discountquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/discount"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const discountListOperation = "billing.catalog.discount.list"

func (runtime *GlobalHTTPRuntime) ListDiscounts(
	ctx context.Context, query billingapi.Query,
) (billingapi.QueryPage, error) {
	capability, err := snapshotCapability(query.Access, discountListOperation)
	if err != nil {
		return billingapi.QueryPage{}, err
	}
	request, err := discountquery.FromTransportedQueryRequest(query.Request, query.Transport)
	if err != nil {
		return billingapi.QueryPage{}, err
	}
	var result discountquery.Result
	var actionErr error
	err = runtime.catalog.Snapshot(ctx, capability,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			result, actionErr = runtime.discounts.Execute(scoped, scope, request)
			return actionErr
		})
	if actionErr != nil {
		return billingapi.QueryPage{}, actionErr
	}
	if err != nil {
		return billingapi.QueryPage{}, billingapi.ErrAuthorityUnavailable
	}
	return toQueryPage(result), nil
}
