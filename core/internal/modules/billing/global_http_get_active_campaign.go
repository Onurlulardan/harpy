package billing

import (
	"context"

	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *GlobalHTTPRuntime) GetActiveCampaign(
	ctx context.Context,
) (*billingapi.CampaignView, error) {
	now := runtime.clock()
	request, err := catalogquery.NewActiveCampaignRequest(now)
	if err != nil {
		return nil, billingapi.ErrAuthorityUnavailable
	}
	groupGeneration, billingGeneration, err :=
		runtime.generations.ResolveGlobalBillingGenerations(ctx)
	if err != nil || groupGeneration < 1 || billingGeneration < 1 {
		return nil, billingapi.ErrAuthorityUnavailable
	}
	var snapshot *catalogquery.CampaignSnapshot
	var actionErr error
	err = runtime.catalog.PublicSnapshot(
		ctx, platformtx.PublicBillingCatalogGet, groupGeneration, billingGeneration,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			store, bindErr := runtime.catalogStore(scope)
			if bindErr != nil {
				actionErr = bindErr
				return bindErr
			}
			queries, buildErr := catalogquery.New(store)
			if buildErr != nil {
				actionErr = buildErr
				return buildErr
			}
			snapshot, actionErr = queries.ReadActiveCampaign(scoped, request)
			return actionErr
		},
	)
	if actionErr != nil || err != nil {
		return nil, billingapi.ErrAuthorityUnavailable
	}
	if snapshot == nil {
		return nil, nil
	}
	view := campaignView(*snapshot)
	return &view, nil
}
