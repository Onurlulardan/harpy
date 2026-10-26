package billing

import (
	campaignpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/campaign"
	discountredemptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/discount"
	discountpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/discountcatalog"
	invoicepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/invoice"
	legalholdpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/legalhold"
	offeringpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/offering"
	usagepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/usage"
	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
	discountredemptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func BindCampaignStore(scope platformtx.GlobalScope) (campaignapp.Store, error) {
	return campaignpostgres.Bind(scope)
}

func BindDiscountCatalogStore(scope platformtx.GlobalScope) (discountapp.Store, error) {
	return discountpostgres.Bind(scope)
}

func BindDiscountRedemptionStore(
	scope platformtx.TenantScope,
) (discountredemptionapp.Store, error) {
	return discountredemptionpostgres.Bind(scope)
}

func BindOfferingStore(scope platformtx.GlobalScope) (offeringapp.Store, error) {
	return offeringpostgres.Bind(scope)
}

func BindInvoiceStore(
	scope platformtx.TenantScope,
	catalogs platformasync.WriteCatalogs,
) (invoiceapp.Store, error) {
	authority, err := platformtx.MintOwnerAuthority(scope, "billing")
	if err != nil || !authority.OwnsScope(scope) {
		return nil, invoiceapp.ErrConfiguration
	}
	effects, err := newInvoiceEffects(scope, authority, catalogs)
	if err != nil {
		return nil, err
	}
	return invoicepostgres.Bind(scope, effects)
}

func BindLegalHoldStore(scope platformtx.TenantScope) (legalholdapp.Store, error) {
	return legalholdpostgres.Bind(scope)
}

func BindUsageStore(
	scope platformtx.TenantScope,
	catalogs platformasync.WriteCatalogs,
) (usageapp.Store, error) {
	authority, err := platformtx.MintOwnerAuthority(scope, "billing")
	if err != nil || !authority.OwnsScope(scope) {
		return nil, usageapp.ErrConfiguration
	}
	effects, err := newUsageEffects(scope, authority, catalogs)
	if err != nil {
		return nil, err
	}
	return usagepostgres.Bind(scope, effects)
}
