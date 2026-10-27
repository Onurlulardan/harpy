package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PermissionRequestRuntime struct {
	service *permissionrequest.Service
	catalog *permissionrequest.CatalogService
}

func NewPermissionRequestRuntime(
	catalogs platformasync.WriteCatalogs,
	now PermissionRequestClock,
) (*PermissionRequestRuntime, error) {
	service, catalog, err := buildPermissionRequestServices(catalogs, now)
	if err != nil {
		return nil, err
	}
	return &PermissionRequestRuntime{service: service, catalog: catalog}, nil
}

func (runtime *PermissionRequestRuntime) Valid() bool {
	return runtime != nil && runtime.service != nil && runtime.catalog != nil
}

func (runtime *PermissionRequestRuntime) decide(
	ctx context.Context,
	scope platformtx.TenantScope,
	command permissionrequest.DecideCommand,
) (permissionrequest.DecisionResult, error) {
	if !runtime.Valid() {
		return permissionrequest.DecisionResult{}, permissionrequest.ErrUnavailable
	}
	return runtime.service.Decide(ctx, scope, command)
}
