package billing

import (
	"context"
	"fmt"
	"time"

	billingapplication "github.com/Onurlulardan/harpy/core/internal/modules/billing/application"
	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
	campaignquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/campaign"
	discountquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/discount"
	offeringquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/offering"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type GlobalHTTPGenerationSource interface {
	ResolveGlobalBillingGenerations(context.Context) (int64, int64, error)
}

type globalHTTPDependencies struct {
	Catalog       *billingapplication.GlobalCatalog
	Generations   GlobalHTTPGenerationSource
	CatalogStore  func(platformtx.GlobalScope) (catalogquery.Store, error)
	OfferingStore func(platformtx.GlobalScope) (offeringapp.Store, error)
	DiscountStore func(platformtx.GlobalScope) (discountapp.Store, error)
	CampaignStore func(platformtx.GlobalScope) (campaignapp.Store, error)
	Offerings     *offeringquery.Service
	Discounts     *discountquery.Service
	Campaigns     *campaignquery.Service
	NextID        func() (uuid.UUID, error)
	Clock         func() time.Time
}

type GlobalHTTPRuntime struct {
	catalog       *billingapplication.GlobalCatalog
	generations   GlobalHTTPGenerationSource
	catalogStore  func(platformtx.GlobalScope) (catalogquery.Store, error)
	offeringStore func(platformtx.GlobalScope) (offeringapp.Store, error)
	discountStore func(platformtx.GlobalScope) (discountapp.Store, error)
	campaignStore func(platformtx.GlobalScope) (campaignapp.Store, error)
	offerings     *offeringquery.Service
	discounts     *discountquery.Service
	campaigns     *campaignquery.Service
	nextID        func() (uuid.UUID, error)
	clock         func() time.Time
}

func newGlobalHTTPRuntime(dependencies globalHTTPDependencies) (*GlobalHTTPRuntime, error) {
	if dependencies.Catalog == nil || dependencies.Generations == nil ||
		dependencies.CatalogStore == nil || dependencies.OfferingStore == nil ||
		dependencies.DiscountStore == nil || dependencies.CampaignStore == nil ||
		dependencies.Offerings == nil || dependencies.Discounts == nil ||
		dependencies.Campaigns == nil || dependencies.NextID == nil ||
		dependencies.Clock == nil {
		return nil, fmt.Errorf("global Billing HTTP dependencies are required")
	}
	return &GlobalHTTPRuntime{
		catalog: dependencies.Catalog, generations: dependencies.Generations,
		catalogStore: dependencies.CatalogStore, offeringStore: dependencies.OfferingStore,
		discountStore: dependencies.DiscountStore, campaignStore: dependencies.CampaignStore,
		offerings: dependencies.Offerings, discounts: dependencies.Discounts,
		campaigns: dependencies.Campaigns, nextID: dependencies.NextID,
		clock: dependencies.Clock,
	}, nil
}

var _ billingapi.PublicQueries = (*GlobalHTTPRuntime)(nil)
var _ billingapi.PlatformCatalog = (*GlobalHTTPRuntime)(nil)
var _ billingapi.LegacyIdentityResolver = (*GlobalHTTPRuntime)(nil)
var _ billingapi.PublicLegacyIdentityResolver = (*GlobalHTTPRuntime)(nil)
