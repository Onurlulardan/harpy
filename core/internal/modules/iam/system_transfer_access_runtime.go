package iam

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type SystemTransferAccessPlatform interface {
	IdentityHTTPPlatform
	tenantSlugResolver
}

type SystemTransferAccessRuntime struct {
	pool        SystemTransferAccessPlatform
	coordinator *platformtx.Coordinator
	security    *HTTPSecurity
	authorizer  authorization.Authorizer
	clock       func() time.Time
}

func NewSystemTransferAccessRuntime(pool SystemTransferAccessPlatform,
	coordinator *platformtx.Coordinator, security *HTTPSecurity,
	clock func() time.Time) (*SystemTransferAccessRuntime, error) {
	if pool == nil || coordinator == nil || security == nil ||
		!security.Valid() || clock == nil {
		return nil, fmt.Errorf("IAM system transfer access dependencies are required")
	}
	authorizer, err := newSystemLayoutAuthorizer()
	if err != nil {
		return nil, err
	}
	return &SystemTransferAccessRuntime{pool: pool, coordinator: coordinator,
		security: security, authorizer: authorizer, clock: clock}, nil
}

var _ SystemTransferAccessVerifier = (*SystemTransferAccessRuntime)(nil)
