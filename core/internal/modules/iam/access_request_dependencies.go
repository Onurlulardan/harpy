package iam

import (
	"fmt"
	"time"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type accessRequestPlatform interface {
	IdentityHTTPPlatform
	tenantSlugResolver
}

type accessRequestDependencies struct {
	pool        accessRequestPlatform
	coordinator *platformtx.Coordinator
	security    *HTTPSecurity
	authorizer  authorization.Authorizer
	clock       func() time.Time
}

func NewSharedAccessDependencies(
	pool accessRequestPlatform,
	coordinator *platformtx.Coordinator,
	security *HTTPSecurity,
	clock func() time.Time,
) (*accessRequestDependencies, error) {
	if pool == nil || coordinator == nil || !security.Valid() || clock == nil {
		return nil, fmt.Errorf("IAM access dependencies are required")
	}
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM access authorization evidence: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build IAM access authorizer: %w", err)
	}
	return &accessRequestDependencies{
		pool: pool, coordinator: coordinator, security: security,
		authorizer: authorizer, clock: clock,
	}, nil
}
