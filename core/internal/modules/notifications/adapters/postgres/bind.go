package postgres

import (
	"context"
	"errors"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const storageOwner = "notifications"

type JobWriterBinder func(
	platformtx.TenantScope,
	platformtx.OwnerAuthority,
) (EnqueueJob, error)

type Binder struct {
	jobs   jobsv1.Catalog
	writer JobWriterBinder
}

func NewBinder(jobs jobsv1.Catalog, writer JobWriterBinder) (*Binder, error) {
	validated, err := jobsv1.NewCatalog(jobs.Policies())
	if err != nil || writer == nil {
		return nil, ports.ErrBinding
	}
	return &Binder{jobs: validated, writer: writer}, nil
}

func (binder *Binder) Bind(ctx context.Context, abstract api.Scope) (ports.Stores, error) {
	scope, ok := abstract.(platformtx.TenantScope)
	if binder == nil || !ok || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, scope.Mode()) != nil {
		return nil, ports.ErrBinding
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *Store {
			return tenantStore(database, scope, binder.jobs)
		})
	if err != nil {
		return nil, ports.ErrBinding
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, ports.ErrBinding
	}
	if scope.Mode() != platformtx.Command {
		return bound, nil
	}
	authority, err := platformtx.MintOwnerAuthority(scope, storageOwner)
	if err != nil || !authority.OwnsScope(scope) {
		return nil, ports.ErrBinding
	}
	bound.authority = authorityStamp{domain: authority.Domain(),
		generation: authority.Generation(), groupID: authority.GroupID(),
		groupGen: authority.GroupGeneration()}
	bound.enqueue, err = binder.writer(scope, authority)
	if err != nil || bound.enqueue == nil {
		return nil, ports.ErrBinding
	}
	return bound, nil
}

func (binder *Binder) Recipient(ctx context.Context, abstract api.Scope,
	recipientID domain.ID,
) (ports.Recipient, error) {
	scope, ok := abstract.(platformtx.TenantScope)
	if binder == nil || !ok || ctx == nil || ctx.Err() != nil ||
		!recipientID.Valid() || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireOwner(scope, storageOwner) != nil {
		return ports.Recipient{}, ErrInvalid
	}
	stores, err := binder.Bind(ctx, scope)
	store, bound := stores.(*Store)
	if err != nil || !bound || store == nil {
		return ports.Recipient{}, ErrInvalid
	}
	resolved, err := scanNotificationRecipient(store.queryRow(
		ctx, resolveNotificationRecipientSQL, recipientID.String()))
	if err != nil || resolved.principalID != recipientID {
		return ports.Recipient{}, ErrInvalid
	}
	tenantUUID := scope.TenantID()
	tenantID, err := domain.IDFromBytes(tenantUUID[:])
	if err != nil {
		return ports.Recipient{}, ErrInvalid
	}
	return ports.Recipient{TenantID: tenantID, PrincipalID: resolved.principalID,
		SupportedLocales: append([]string(nil), resolved.supportedLocales...),
		EmailVerified:    resolved.emailVerified, TelegramVerified: resolved.telegramVerified,
		AuthorizationVersion: resolved.authorizationVersion}, nil
}

func tenantStore(database platformpostgres.BoundExecutor, scope platformtx.TenantScope,
	jobs jobsv1.Catalog,
) *Store {
	return &Store{command: scope.Mode() == platformtx.Command,
		tenantID: scope.TenantID(), jobs: jobs,
		queryRow: func(ctx context.Context, sql string, arguments ...any) row {
			return database.QueryRow(ctx, sql, arguments...)
		},
		query: func(ctx context.Context, sql string, arguments ...any) (rows, error) {
			return database.Query(ctx, sql, arguments...)
		},
		execute: func(ctx context.Context, sql string, arguments ...any) (commandResult, error) {
			return database.Execute(ctx, sql, arguments...)
		},
		notFound: platformpostgres.RowNotFound,
		allowed: func(operation context.Context) bool {
			return operation != nil && operation.Err() == nil &&
				!platformtx.NetworkAllowed(operation) &&
				platformtx.ContextOwnsScope(operation, scope) &&
				platformtx.RequireActive(scope, scope.Mode()) == nil
		}}
}

type DeliveryRuntime struct {
	withStore func(context.Context, uuid.UUID, func(context.Context, *Store) error) error
}

func NewDeliveryRuntime(coordinator *platformtx.Coordinator) (*DeliveryRuntime, error) {
	if coordinator == nil {
		return nil, ErrInvalid
	}
	binder, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformtx.JobBoundExecutor) *Store { return jobStore(database) })
	if err != nil {
		return nil, errors.Join(ErrInvalid, err)
	}
	runtime := &DeliveryRuntime{}
	runtime.withStore = func(ctx context.Context, tenant uuid.UUID,
		operation func(context.Context, *Store) error,
	) error {
		if ctx == nil || tenant == uuid.Nil || operation == nil {
			return ErrInvalid
		}
		return coordinator.InJobTenant(ctx, tenant,
			func(scoped context.Context, scope platformtx.JobTenantScope) error {
				if !platformtx.ContextOwnsJobScope(scoped, scope) ||
					platformtx.NetworkAllowed(scoped) || scope.TenantID() != tenant {
					return ErrInvalid
				}
				store, bindErr := platformtx.BindJob(scope, binder)
				if bindErr != nil || store == nil {
					return errors.Join(ErrInvalid, bindErr)
				}
				return operation(scoped, store)
			})
	}
	return runtime, nil
}

func jobStore(database platformtx.JobBoundExecutor) *Store {
	return &Store{command: true,
		queryRow: func(ctx context.Context, sql string, arguments ...any) row {
			return database.QueryRow(ctx, sql, arguments...)
		},
		query: func(ctx context.Context, sql string, arguments ...any) (rows, error) {
			return database.Query(ctx, sql, arguments...)
		},
		execute: func(ctx context.Context, sql string, arguments ...any) (commandResult, error) {
			return database.Execute(ctx, sql, arguments...)
		},
		notFound: platformpostgres.RowNotFound,
		allowed: func(ctx context.Context) bool {
			return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx)
		}}
}

var _ ports.Binder = (*Binder)(nil)
var _ ports.DeliveryRuntimeStore = (*DeliveryRuntime)(nil)

func IsRowNotFound(err error) bool { return platformpostgres.RowNotFound(err) }
