package billing

import (
	"context"

	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *GlobalHTTPRuntime) readCampaign(
	ctx context.Context, capability platformtx.GlobalSnapshotCapability, id uuid.UUID,
) (billingapi.CampaignView, error) {
	var snapshot catalogquery.CampaignSnapshot
	var actionErr error
	err := runtime.catalog.Snapshot(ctx, capability,
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
			resolved, resolveErr := queries.ResolvePlatformCatalog(
				scoped, catalogquery.PlatformCampaign, id,
			)
			if resolveErr != nil {
				actionErr = resolveErr
				return resolveErr
			}
			var valid bool
			snapshot, valid = resolved.Campaign()
			if !valid {
				actionErr = catalogquery.ErrResult
			}
			return actionErr
		})
	if actionErr != nil {
		return billingapi.CampaignView{}, normalizeCatalogError(actionErr)
	}
	if err != nil {
		return billingapi.CampaignView{}, billingapi.ErrAuthorityUnavailable
	}
	return campaignView(snapshot), nil
}
