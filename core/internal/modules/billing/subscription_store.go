package billing

import (
	subscriptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/subscription"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func BindSubscriptionActorStore(
	scope platformtx.TenantScope,
	catalogs platformasync.WriteCatalogs,
) (subscriptionapp.Store, error) {
	authority, err := platformtx.MintOwnerAuthority(scope, "billing")
	if err != nil || !authority.OwnsScope(scope) {
		return nil, subscriptionapp.ErrConfiguration
	}
	effects, err := newSubscriptionEffects(scope, authority, catalogs)
	if err != nil {
		return nil, err
	}
	return subscriptionpostgres.BindActor(scope, effects)
}
