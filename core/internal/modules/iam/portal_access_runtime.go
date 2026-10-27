package iam

import (
	"fmt"
	"time"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	portaluserpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/portaluser"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PortalAccessPlatform interface {
	IdentityHTTPPlatform
	tenantSlugResolver
}

type PortalAccessRuntime struct {
	pool        PortalAccessPlatform
	coordinator *platformtx.Coordinator
	security    *HTTPSecurity
	users       *portaluser.Service
	authorizer  authorization.Authorizer
	clock       func() time.Time
}

func NewPortalAccessRuntime(
	pool PortalAccessPlatform, coordinator *platformtx.Coordinator,
	security *HTTPSecurity, clock func() time.Time,
) (*PortalAccessRuntime, error) {
	if pool == nil || coordinator == nil || !security.Valid() || clock == nil {
		return nil, fmt.Errorf("IAM portal access dependencies are required")
	}
	store, err := portaluserpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	users, err := portaluser.NewService(store)
	if err != nil {
		return nil, err
	}
	port, err := authorizationpostgres.NewPortalAuthorizer()
	if err != nil {
		return nil, err
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, err
	}
	return &PortalAccessRuntime{pool: pool, coordinator: coordinator,
		security: security, users: users, authorizer: authorizer, clock: clock}, nil
}

func (runtime *PortalAccessRuntime) Ready() bool {
	return runtime != nil && runtime.pool != nil && runtime.coordinator != nil &&
		runtime.security.Valid() && runtime.users != nil && runtime.clock != nil
}
