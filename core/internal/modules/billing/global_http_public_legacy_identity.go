package billing

import (
	"context"

	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *GlobalHTTPRuntime) ResolvePublicLegacyIdentity(
	ctx context.Context,
	kind billingapi.LegacyIdentityKind,
	rawID string,
) (billingapi.LegacyIdentityView, error) {
	resolvedKind, _, valid := globalLegacyIdentityKind(kind)
	identifier, idErr := uuid.Parse(rawID)
	if !valid || idErr != nil || identifier == uuid.Nil {
		return billingapi.LegacyIdentityView{}, billingapi.ErrInvalid
	}
	group, domain, err := runtime.generations.ResolveGlobalBillingGenerations(ctx)
	if err != nil || group < 1 || domain < 1 {
		return billingapi.LegacyIdentityView{}, billingapi.ErrAuthorityUnavailable
	}
	var record catalogquery.LegacyIdentityRecord
	var actionErr error
	err = runtime.catalog.PublicSnapshot(
		ctx, platformtx.PublicBillingCatalogList, group, domain,
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
			record, actionErr = queries.MaterializeLegacyIdentity(
				scoped, resolvedKind, identifier,
			)
			return actionErr
		},
	)
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
