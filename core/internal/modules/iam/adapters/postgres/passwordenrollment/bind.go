package passwordenrollmentpostgres

import (
	"context"

	passwordenrollment "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/passwordenrollment"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type tenantScope = platformtx.TenantScope
type scopeMode = platformtx.ScopeMode

const commandMode scopeMode = platformtx.Command

type boundStore struct {
	database platformpostgres.BoundExecutor
}

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (passwordenrollment.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, passwordenrollment.ErrStorage
	}
	return &store{binder: binder, valid: true}, nil
}

func (storage *store) bind(
	scope tenantScope,
) (*boundStore, error) {
	if storage == nil || !storage.valid {
		return nil, passwordenrollment.ErrStorage
	}
	bound, err := platformtx.Bind(scope, storage.binder)
	if err != nil || bound == nil {
		return nil, passwordenrollment.ErrStorage
	}
	return bound, nil
}

var _ passwordenrollment.Store = (*store)(nil)

func scopeUsable(ctx context.Context, scope tenantScope, mode scopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

func databaseID(id passwordenrollment.IdentityID) uuid.UUID { return uuid.UUID(id) }
