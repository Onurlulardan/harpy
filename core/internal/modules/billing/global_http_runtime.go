package billing

import (
	"fmt"
	"time"

	executionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/restquery/execution"
	billingapplication "github.com/Onurlulardan/harpy/core/internal/modules/billing/application"
	campaignquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/campaign"
	discountquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/discount"
	offeringquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/offering"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func NewGlobalHTTPRuntime(
	generations GlobalHTTPGenerationSource,
	coordinator *platformtx.Coordinator,
	keyring config.CursorKeyring,
) (*GlobalHTTPRuntime, error) {
	if nilPort(generations) || coordinator == nil {
		return nil, fmt.Errorf("global Billing HTTP platform dependencies are required")
	}
	catalog, err := billingapplication.NewGlobalCatalog(coordinator)
	if err != nil {
		return nil, err
	}
	keys, err := globalHTTPQueryKeys(keyring)
	if err != nil {
		return nil, err
	}
	owner, err := executionpostgres.New()
	if err != nil {
		return nil, err
	}
	offerings, err := offeringquery.NewService(owner, keys)
	if err != nil {
		return nil, err
	}
	discounts, err := discountquery.NewService(owner, keys)
	if err != nil {
		return nil, err
	}
	campaigns, err := campaignquery.NewService(owner, keys)
	if err != nil {
		return nil, err
	}
	clock := func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }
	return newGlobalHTTPRuntime(globalHTTPDependencies{
		Catalog: catalog, Generations: generations, CatalogStore: BindCatalogQueryStore,
		OfferingStore: BindOfferingStore, DiscountStore: BindDiscountCatalogStore,
		CampaignStore: BindCampaignStore, Offerings: offerings,
		Discounts: discounts, Campaigns: campaigns, NextID: uuid.NewV7, Clock: clock,
	})
}
