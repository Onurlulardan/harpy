package recoveryanchorpostgres

import (
	"context"
	"fmt"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	anchorports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/recoveryanchor"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantScope = platformtx.TenantScope
type scopeMode = platformtx.ScopeMode
type AuthorityFinalizerPool = platformpostgres.Pool

const (
	commandMode  scopeMode = platformtx.Command
	snapshotMode scopeMode = platformtx.SnapshotQuery
)

type scanner interface{ Scan(...any) error }
type boundStore struct {
	database platformpostgres.BoundExecutor
}
type store struct {
	binder      platformpostgres.OwnerBinder[*boundStore]
	rowNotFound func(error) bool
}

func NewStore() (anchorports.Store[tenantScope], error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, anchordomain.ErrDependency
	}
	return &store{binder: binder, rowNotFound: platformpostgres.RowNotFound}, nil
}

func NewAuthorityFinalizerStore(
	pool *AuthorityFinalizerPool,
) (anchorports.AuthorityFinalizerStore, error) {
	return platformpostgres.BindIAMAuthorityFinalizer(
		pool,
		func(
			transactions platformpostgres.IAMAuthorityFinalizerTransactions,
		) (anchorports.AuthorityFinalizerStore, error) {
			if transactions == nil {
				return nil, fmt.Errorf("IAM authority finalizer is unavailable")
			}
			return &authorityFinalizerStore{
				transactions: platformAuthorityFinalizerTransactions{
					value: transactions,
				},
			}, nil
		},
	)
}

type platformAuthorityFinalizerTransactions struct {
	value platformpostgres.IAMAuthorityFinalizerTransactions
}

func (transactions platformAuthorityFinalizerTransactions) Within(
	ctx context.Context,
	callback func(authorityFinalizerDatabase) error,
) error {
	return transactions.value.Within(
		ctx,
		platformpostgres.TransactionCommand,
		func(database platformpostgres.IAMAuthorityFinalizerDatabase) error {
			return callback(platformAuthorityFinalizerDatabase{value: database})
		},
	)
}

type platformAuthorityFinalizerDatabase struct {
	value platformpostgres.IAMAuthorityFinalizerDatabase
}

func (database platformAuthorityFinalizerDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) authorityFinalizerScanner {
	return database.value.QueryRow(ctx, statement, arguments...)
}

func (storage *store) bind(
	ctx context.Context,
	scope tenantScope,
	mode scopeMode,
) (*boundStore, error) {
	if storage == nil || ctx == nil || ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) || platformtx.RequireActive(scope, mode) != nil {
		return nil, anchordomain.ErrInvalid
	}
	bound, err := platformtx.Bind(scope, storage.binder)
	if err != nil || bound == nil {
		return nil, anchordomain.ErrDependency
	}
	return bound, nil
}

var _ anchorports.Store[tenantScope] = (*store)(nil)
var _ anchorports.AuthorityFinalizerStore = (*authorityFinalizerStore)(nil)
