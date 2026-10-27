package iam

import (
	"fmt"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type permissionRequestHTTPApplication struct {
	pool        PermissionRequestHTTPPlatform
	coordinator *platformtx.Coordinator
	runtime     *PermissionRequestRuntime
	authorizer  authorization.Authorizer
}

func NewPermissionRequestHTTPApplication(
	pool PermissionRequestHTTPPlatform,
	coordinator *platformtx.Coordinator,
	runtime *PermissionRequestRuntime,
) (PermissionRequestHTTPApplication, error) {
	if pool == nil || coordinator == nil || !runtime.Valid() {
		return nil, fmt.Errorf("IAM permission request HTTP dependencies are required")
	}
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build permission request authorizer: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build permission request policy: %w", err)
	}
	return &permissionRequestHTTPApplication{
		pool: pool, coordinator: coordinator, runtime: runtime, authorizer: authorizer,
	}, nil
}
