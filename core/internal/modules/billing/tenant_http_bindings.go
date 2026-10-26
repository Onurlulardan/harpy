package billing

import (
	"time"

	discountpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/discount"
	eligibilitypostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/discounteligibility"
	subscriptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/subscription"
	usagepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/usage"
	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type tenantHTTPBindings struct {
	catalogs platformasync.WriteCatalogs
	nextID   func() (uuid.UUID, error)
	clock    func() time.Time
}

func (bindings *tenantHTTPBindings) SubscriptionQueries(
	scope platformtx.TenantScope,
) (billingapi.TenantSubscriptionQueryPort, error) {
	store, err := subscriptionpostgres.BindQueries(scope)
	if err != nil {
		return nil, err
	}
	queries, err := subscriptionapp.NewQueries(store)
	if err != nil {
		return nil, err
	}
	return &tenantSubscriptionQueries{queries: queries}, nil
}

func (bindings *tenantHTTPBindings) UsageQueries(
	scope platformtx.TenantScope,
) (billingapi.TenantUsageQueryPort, error) {
	store, err := usagepostgres.BindQueries(scope)
	if err != nil {
		return nil, err
	}
	queries, err := usageapp.NewQueries(store)
	if err != nil {
		return nil, err
	}
	return &tenantUsageQueries{queries: queries}, nil
}

func (bindings *tenantHTTPBindings) DiscountQueries(
	scope platformtx.TenantScope,
) (billingapi.TenantDiscountQueryPort, error) {
	store, err := eligibilitypostgres.Bind(scope)
	if err != nil {
		return nil, err
	}
	queries, err := discountapp.NewQueries(store, bindings.clock)
	if err != nil {
		return nil, err
	}
	return &tenantDiscountQueries{queries: queries}, nil
}

func (bindings *tenantHTTPBindings) SubscriptionCommands(
	scope platformtx.TenantScope,
) (billingapi.TenantSubscriptionCommandPort, error) {
	store, err := BindSubscriptionActorStore(scope, bindings.catalogs)
	if err != nil {
		return nil, err
	}
	commands, err := subscriptionapp.NewCommands(store, bindings.nextID, bindings.clock)
	if err != nil {
		return nil, err
	}
	return &tenantSubscriptionCommands{commands: commands}, nil
}

func (bindings *tenantHTTPBindings) DiscountCommands(
	scope platformtx.TenantScope,
) (billingapi.TenantDiscountCommandPort, error) {
	store, err := discountpostgres.Bind(scope)
	if err != nil {
		return nil, err
	}
	commands, err := discountapp.NewCommands(store, bindings.nextID, bindings.clock)
	if err != nil {
		return nil, err
	}
	return &tenantDiscountCommands{commands: commands}, nil
}
