package catalogquery

import (
	"context"
	"fmt"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func (storage *store) ResolvePlatformCatalog(
	ctx context.Context, query catalogapp.PlatformResolveStoreQuery,
) (catalogapp.PlatformCatalogRecord, bool, error) {
	if !storage.usable(ctx) {
		return catalogapp.PlatformCatalogRecord{}, false, catalogapp.ErrConfiguration
	}
	switch query.Kind {
	case catalogapp.PlatformOffering:
		record, found, err := storage.platformOffering(ctx, query.ID)
		if err != nil || !found {
			return catalogapp.PlatformCatalogRecord{}, found, err
		}
		return catalogapp.PlatformCatalogRecord{Offering: &record}, true, nil
	case catalogapp.PlatformDiscount:
		record, found, err := storage.platformDiscount(ctx, query.ID)
		if err != nil || !found {
			return catalogapp.PlatformCatalogRecord{}, found, err
		}
		return catalogapp.PlatformCatalogRecord{Discount: &record}, true, nil
	case catalogapp.PlatformCampaign:
		record, found, err := storage.platformCampaign(ctx, query.ID)
		if err != nil || !found {
			return catalogapp.PlatformCatalogRecord{}, found, err
		}
		return catalogapp.PlatformCatalogRecord{Campaign: &record}, true, nil
	default:
		return catalogapp.PlatformCatalogRecord{}, false,
			fmt.Errorf("resolve platform billing catalog: unsupported kind")
	}
}
