package iam

import (
	"fmt"
	"time"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type rbacHTTPApplication struct {
	pool        IdentityHTTPPlatform
	coordinator *platformtx.Coordinator
	runtime     *RBACRuntime
	authorizer  authorization.Authorizer
	now         RBACClock
}

func NewRBACHTTPApplication(
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	runtime *RBACRuntime,
) (RBACHTTPApplication, error) {
	return newRBACHTTPApplication(pool, coordinator, runtime, time.Now)
}

func newRBACHTTPApplication(
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	runtime *RBACRuntime,
	now RBACClock,
) (RBACHTTPApplication, error) {
	if pool == nil || coordinator == nil || !runtime.Valid() || now == nil {
		return nil, fmt.Errorf("IAM RBAC HTTP dependencies are required")
	}
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM RBAC authorizer: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build IAM RBAC policy: %w", err)
	}
	return &rbacHTTPApplication{
		pool: pool, coordinator: coordinator, runtime: runtime,
		authorizer: authorizer, now: now,
	}, nil
}
