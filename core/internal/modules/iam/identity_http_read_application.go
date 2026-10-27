package iam

import (
	"fmt"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// NewIdentityHTTPReadApplication builds the identity query graph.
// collection reads. It does not load API token, password or directory secrets.
func NewIdentityHTTPReadApplication(
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	identities *IdentityRuntime,
	queries *IAMRESTQueryRuntime,
	cursorKeys config.CursorKeyring,
) (IdentityHTTPApplication, error) {
	if pool == nil || coordinator == nil || !identities.Valid() || !queries.Valid() {
		return nil, fmt.Errorf("IAM identity read application dependencies are required")
	}
	authorizationPort, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM identity read authorizer: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(authorizationPort)
	if err != nil {
		return nil, fmt.Errorf("build IAM identity read policy: %w", err)
	}
	queryKeys, err := identityQueryKeysFromConfig(cursorKeys)
	if err != nil {
		return nil, err
	}
	queryService, err := queries.newSystemResourceQuery(queryKeys)
	if err != nil {
		return nil, err
	}
	return &identityHTTPApplication{
		pool: pool, coordinator: coordinator, identities: identities,
		queries: queryService, authorizer: authorizer,
	}, nil
}
