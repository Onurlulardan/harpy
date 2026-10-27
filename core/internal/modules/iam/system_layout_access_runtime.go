package iam

import (
	"fmt"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type SystemLayoutAccessPlatform interface {
	IdentityHTTPPlatform
	tenantSlugResolver
}

type SystemLayoutAccessRuntime struct {
	pool        SystemLayoutAccessPlatform
	coordinator *platformtx.Coordinator
	security    *HTTPSecurity
	authorizer  authorization.Authorizer
	clock       func() time.Time
}

func NewSystemLayoutAccessRuntime(
	pool SystemLayoutAccessPlatform,
	coordinator *platformtx.Coordinator,
	security *HTTPSecurity,
	clock func() time.Time,
) (*SystemLayoutAccessRuntime, error) {
	if pool == nil || coordinator == nil || security == nil ||
		!security.Valid() || clock == nil {
		return nil, fmt.Errorf("IAM system layout access dependencies are required")
	}
	authorizer, err := newSystemLayoutAuthorizer()
	if err != nil {
		return nil, err
	}
	return &SystemLayoutAccessRuntime{
		pool: pool, coordinator: coordinator, security: security,
		authorizer: authorizer, clock: clock,
	}, nil
}

func systemLayoutAdmission(
	tenantID, actorID publicresource.ID,
	tokenVersion, authorizationVersion, generation int64,
) (publicresource.Admission, error) {
	actor, err := publicresource.NewActorEvidence(
		tenantID, actorID, tokenVersion, authorizationVersion,
	)
	if err != nil {
		return publicresource.Admission{}, err
	}
	return publicresource.NewAdmission(tenantID, actor, generation, generation, generation)
}
