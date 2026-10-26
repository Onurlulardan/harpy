package iam

import (
	"context"
	"fmt"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	portaluserpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/portaluser"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type portalSetupStarter interface {
	StartPortalSetup(
		ctx context.Context, claims accesstoken.Claims, input LifecycleResourceInput,
	) error
}

type portalUserHTTPApplication struct {
	pool        IdentityHTTPPlatform
	coordinator *platformtx.Coordinator
	identities  *IdentityRuntime
	resolver    PortalSubjectResolver
	users       *portaluser.Service
	authorizer  authorization.Authorizer
	directory   identityDirectoryKey
	changes     IdentityMutationSink
	setups      portalSetupStarter
}

func NewPortalUserHTTPApplication(
	pool IdentityHTTPPlatform, coordinator *platformtx.Coordinator,
	identities *IdentityRuntime, resolver PortalSubjectResolver,
	settings config.IAMAPI, changes IdentityMutationSink,
	setups LifecycleHTTPApplication,
) (PortalUserHTTPApplication, error) {
	if pool == nil || coordinator == nil || !identities.Valid() || resolver == nil ||
		nilIdentityDependency(changes) || setups == nil {
		return nil, fmt.Errorf("IAM portal user application dependencies are required")
	}
	store, err := portaluserpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	users, err := portaluser.NewService(store)
	if err != nil {
		return nil, err
	}
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, err
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, err
	}
	directory, err := identityDirectoryFromConfig(settings)
	if err != nil {
		return nil, err
	}
	return &portalUserHTTPApplication{pool: pool, coordinator: coordinator,
		identities: identities, resolver: resolver, users: users,
		authorizer: authorizer, directory: directory, changes: changes, setups: setups}, nil
}
