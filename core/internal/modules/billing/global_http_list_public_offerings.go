package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	offeringquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/offering"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *GlobalHTTPRuntime) ListPublicOfferings(
	ctx context.Context, query billingapi.PublicQuery,
) (billingapi.QueryPage, error) {
	request, err := offeringquery.FromTransportedQueryRequest(query.Request, query.Transport)
	if err != nil {
		return billingapi.QueryPage{}, err
	}
	groupGeneration, billingGeneration, err :=
		runtime.generations.ResolveGlobalBillingGenerations(ctx)
	if err != nil || groupGeneration < 1 || billingGeneration < 1 {
		return billingapi.QueryPage{}, billingapi.ErrAuthorityUnavailable
	}
	var result offeringquery.Result
	var actionErr error
	err = runtime.catalog.PublicSnapshot(
		ctx, platformtx.PublicBillingCatalogList, groupGeneration, billingGeneration,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			result, actionErr = runtime.offerings.Execute(scoped, scope, request)
			return actionErr
		},
	)
	if actionErr != nil {
		return billingapi.QueryPage{}, actionErr
	}
	if err != nil {
		return billingapi.QueryPage{}, billingapi.ErrAuthorityUnavailable
	}
	return toQueryPage(result), nil
}
