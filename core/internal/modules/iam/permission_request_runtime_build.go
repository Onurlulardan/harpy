package iam

import (
	permissionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/permissionrequest"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func buildPermissionRequestServices(
	catalogs platformasync.WriteCatalogs,
	now PermissionRequestClock,
) (*permissionrequest.Service, *permissionrequest.CatalogService, error) {
	store, err := permissionpostgres.NewPermissionRequestStore()
	if err != nil {
		return nil, nil, err
	}
	sink, err := NewPermissionRequestDecisionSink(catalogs, now)
	if err != nil {
		return nil, nil, err
	}
	service, err := permissionrequest.NewService(store, sink)
	if err != nil {
		return nil, nil, err
	}
	catalogStore, err := permissionpostgres.NewPermissionRequestCatalogStore()
	if err != nil {
		return nil, nil, err
	}
	catalog, err := permissionrequest.NewCatalogService(catalogStore)
	if err != nil {
		return nil, nil, err
	}
	return service, catalog, nil
}
