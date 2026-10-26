package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	offeringquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/offering"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const offeringListOperation = "billing.catalog.offering.list"

func (runtime *GlobalHTTPRuntime) ListOfferings(
	ctx context.Context, query billingapi.Query,
) (billingapi.QueryPage, error) {
	capability, err := snapshotCapability(query.Access, offeringListOperation)
	if err != nil {
		return billingapi.QueryPage{}, err
	}
	request, err := offeringquery.FromTransportedQueryRequest(query.Request, query.Transport)
	if err != nil {
		return billingapi.QueryPage{}, err
	}
	var result offeringquery.Result
	var actionErr error
	err = runtime.catalog.Snapshot(ctx, capability,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			result, actionErr = runtime.offerings.Execute(scoped, scope, request)
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
