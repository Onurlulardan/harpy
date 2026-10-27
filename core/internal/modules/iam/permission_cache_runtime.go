package iam

import (
	"fmt"

	permissioncachepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/permissioncache"
	permissioncacheapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissioncache"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PermissionCacheRuntime struct {
	pool        IdentityHTTPPlatform
	coordinator *platformtx.Coordinator
	application *permissioncacheapp.Application
}

func NewPermissionCacheRuntime(
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
) (*PermissionCacheRuntime, error) {
	if pool == nil || coordinator == nil {
		return nil, fmt.Errorf("IAM permission cache dependencies are required")
	}
	store, err := permissioncachepostgres.New()
	if err != nil {
		return nil, err
	}
	application, err := permissioncacheapp.New(store)
	if err != nil {
		return nil, err
	}
	return &PermissionCacheRuntime{pool: pool, coordinator: coordinator,
		application: application}, nil
}
