package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	campaignquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/campaign"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const campaignListOperation = "billing.catalog.campaign.list"

func (runtime *GlobalHTTPRuntime) ListCampaigns(
	ctx context.Context, query billingapi.Query,
) (billingapi.QueryPage, error) {
	capability, err := snapshotCapability(query.Access, campaignListOperation)
	if err != nil {
		return billingapi.QueryPage{}, err
	}
	request, err := campaignquery.FromTransportedQueryRequest(query.Request, query.Transport)
	if err != nil {
		return billingapi.QueryPage{}, err
	}
	var result campaignquery.Result
	var actionErr error
	err = runtime.catalog.Snapshot(ctx, capability,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			result, actionErr = runtime.campaigns.Execute(scoped, scope, request)
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
