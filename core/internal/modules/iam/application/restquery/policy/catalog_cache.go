package policy

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type catalogCache struct {
	service *Service
	ctx     context.Context
	scope   transaction.TenantScope
	values  map[uuid.UUID]catalog.Published
}

func newCatalogCache(
	service *Service,
	ctx context.Context,
	scope transaction.TenantScope,
) *catalogCache {
	return &catalogCache{
		service: service, ctx: ctx, scope: scope,
		values: make(map[uuid.UUID]catalog.Published, 3),
	}
}

func (cache *catalogCache) load(
	reference api.CatalogReference,
) (catalog.Published, error) {
	resourceID, err := reference.ResourceID()
	if err != nil {
		return catalog.Published{}, catalog.ErrContract
	}
	if published, found := cache.values[resourceID]; found {
		if !published.Matches(reference) {
			return catalog.Published{}, catalog.ErrContract
		}
		return published, nil
	}
	resource, err := catalog.ResourceByID(resourceID)
	if err != nil {
		return catalog.Published{}, catalog.ErrContract
	}
	published, err := cache.service.catalogs.Catalog(cache.ctx, cache.scope, resource.Kind())
	if err != nil || published.Validate() != nil || !published.Matches(reference) {
		return catalog.Published{}, catalog.ErrContract
	}
	cache.values[resourceID] = published
	return published, nil
}
