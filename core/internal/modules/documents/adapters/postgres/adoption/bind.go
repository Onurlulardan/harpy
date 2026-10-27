package adoption

import (
	"context"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const storageOwner = "documents"

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context,
	scope platformtx.TenantScope,
) (adoptionapp.Store, error) {
	tenantID, err := domain.ParseID(scope.TenantID().String())
	if binder == nil || err != nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil {
		return nil, adoptionapp.ErrUnavailable
	}
	owner, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *Store {
			return &Store{tenantID: tenantID,
				queryRow: func(ctx context.Context, statement string, args ...any) Row {
					return database.QueryRow(ctx, statement, args...)
				},
				execute: func(ctx context.Context, statement string, args ...any) error {
					_, executeErr := database.Execute(ctx, statement, args...)
					return executeErr
				},
				allowed: func(value context.Context) bool {
					return !platformtx.NetworkAllowed(value) &&
						platformtx.ContextOwnsScope(value, scope)
				},
			}
		})
	if err != nil {
		return nil, adoptionapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, owner)
	if err != nil || bound == nil {
		return nil, adoptionapp.ErrUnavailable
	}
	return bound, nil
}

func (binder *Binder) BindJob(ctx context.Context,
	scope platformtx.JobTenantScope,
) (adoptionapp.Store, error) {
	tenantID, err := domain.ParseID(scope.TenantID().String())
	if binder == nil || err != nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsJobScope(ctx, scope) {
		return nil, adoptionapp.ErrUnavailable
	}
	owner, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *Store {
			return &Store{tenantID: tenantID,
				queryRow: func(ctx context.Context, statement string, args ...any) Row {
					return database.QueryRow(ctx, statement, args...)
				},
				execute: func(ctx context.Context, statement string, args ...any) error {
					_, executeErr := database.Execute(ctx, statement, args...)
					return executeErr
				},
				allowed: func(value context.Context) bool {
					return !platformtx.NetworkAllowed(value) &&
						platformtx.ContextOwnsJobScope(value, scope)
				},
			}
		})
	if err != nil {
		return nil, adoptionapp.ErrUnavailable
	}
	bound, err := platformtx.BindJob(scope, owner)
	if err != nil || bound == nil {
		return nil, adoptionapp.ErrUnavailable
	}
	return bound, nil
}

var _ adoptionapp.JobStoreBinder = (*Binder)(nil)
var _ adoptionapp.ScopeStoreBinder = (*Binder)(nil)
