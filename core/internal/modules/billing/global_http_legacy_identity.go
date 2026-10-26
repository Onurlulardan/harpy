package billing

import (
	"context"

	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *GlobalHTTPRuntime) ResolveLegacyIdentity(
	ctx context.Context,
	query billingapi.LegacyIdentityQuery,
) (billingapi.LegacyIdentityView, error) {
	kind, operation, valid := globalLegacyIdentityKind(query.Kind)
	identifier, idErr := uuid.Parse(query.LegacyID)
	if !valid || idErr != nil || identifier == uuid.Nil {
		return billingapi.LegacyIdentityView{}, billingapi.ErrInvalid
	}
	capability, err := catalogReadbackCapability(query.Access, operation)
	if err != nil {
		return billingapi.LegacyIdentityView{}, err
	}
	return runtime.resolveGlobalLegacyIdentity(ctx, capability, kind, identifier)
}

func (runtime *GlobalHTTPRuntime) resolveGlobalLegacyIdentity(
	ctx context.Context,
	capability platformtx.GlobalSnapshotCapability,
	kind catalogquery.PlatformCatalogKind,
	identifier uuid.UUID,
) (billingapi.LegacyIdentityView, error) {
	var record catalogquery.LegacyIdentityRecord
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
			record, actionErr = queries.MaterializeLegacyIdentity(scoped, kind, identifier)
			return actionErr
		})
	if actionErr != nil {
		return billingapi.LegacyIdentityView{}, normalizeCatalogError(actionErr)
	}
	if err != nil {
		return billingapi.LegacyIdentityView{}, billingapi.ErrAuthorityUnavailable
	}
	return billingapi.LegacyIdentityView{
		LegacyID: record.LegacyID.String(), CanonicalID: record.CanonicalID.String(),
	}, nil
}

func globalLegacyIdentityKind(
	kind billingapi.LegacyIdentityKind,
) (catalogquery.PlatformCatalogKind, string, bool) {
	switch kind {
	case billingapi.LegacyIdentityOffering:
		return catalogquery.PlatformOffering, offeringReadOperation, true
	case billingapi.LegacyIdentityDiscount:
		return catalogquery.PlatformDiscount, discountReadOperation, true
	case billingapi.LegacyIdentityCampaign:
		return catalogquery.PlatformCampaign, campaignReadOperation, true
	default:
		return "", "", false
	}
}
