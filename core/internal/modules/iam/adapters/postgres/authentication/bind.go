package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type boundStore struct {
	database platformpostgres.BoundExecutor
}

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (authentication.Port, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, errAuthenticationStorage
	}
	return &store{binder: binder, valid: true}, nil
}

var _ authentication.Port = (*store)(nil)
var _ authentication.InitialPasswordPort = (*store)(nil)

func (store *store) bind(
	ctx context.Context,
	scope authentication.CredentialScope,
	purpose platformtx.CredentialPurpose,
	locator platformtx.CredentialLocatorKind,
) (*boundStore, error) {
	if store == nil || !store.valid || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsCredentialScope(ctx, scope) ||
		platformtx.RequireCredentialPurpose(scope, purpose) != nil ||
		scope.LocatorKind() != locator {
		return nil, errAuthenticationStorage
	}
	bound, err := platformtx.BindCredential(scope, store.binder)
	if err != nil || bound == nil {
		return nil, errAuthenticationStorage
	}
	return bound, nil
}

func (store *store) bindLogin(
	ctx context.Context,
	scope authentication.CredentialScope,
) (*boundStore, error) {
	return store.bind(
		ctx, scope, platformtx.CredentialLogin, platformtx.CredentialLoginDirectory,
	)
}

func (store *store) bindRefresh(
	ctx context.Context,
	scope authentication.CredentialScope,
) (*boundStore, error) {
	return store.bind(
		ctx, scope, platformtx.CredentialRefresh, platformtx.CredentialRefreshLocator,
	)
}

func (store *store) bindReauthentication(
	ctx context.Context,
	scope authentication.CredentialScope,
) (*boundStore, error) {
	return store.bind(
		ctx, scope, platformtx.CredentialReauthentication,
		platformtx.CredentialSessionLocator,
	)
}

func (store *store) bindRecovery(
	ctx context.Context,
	scope authentication.CredentialScope,
) (*boundStore, error) {
	return store.bind(
		ctx, scope, platformtx.CredentialPasswordReset,
		platformtx.CredentialTenantRoute,
	)
}

func (store *store) bindInitialLogin(
	ctx context.Context,
	scope authentication.CredentialScope,
) (*boundStore, error) {
	return store.bind(ctx, scope, platformtx.CredentialInitialPassword,
		platformtx.CredentialLoginDirectory)
}

func (store *store) bindInitialRoute(
	ctx context.Context,
	scope authentication.CredentialScope,
) (*boundStore, error) {
	return store.bind(ctx, scope, platformtx.CredentialInitialPassword,
		platformtx.CredentialTenantRoute)
}
