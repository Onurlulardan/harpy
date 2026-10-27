package publicexecution

import (
	"context"
	"errors"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var ErrResolve = errors.New("analytics public execution root resolution failed")

type row interface{ Scan(...any) error }
type boundResolver struct {
	database platformpostgres.BoundExecutor
}

type Roots interface {
	QueryRoot(context.Context, platformtx.TenantScope,
		uuid.UUID, uuid.UUID) (uuid.UUID, error)
	DashboardRoot(context.Context, platformtx.TenantScope,
		uuid.UUID, uuid.UUID) (uuid.UUID, error)
}

type rootResolver func(context.Context, platformtx.TenantScope,
	uuid.UUID, uuid.UUID) (uuid.UUID, error)

type roots struct {
	query, dashboard rootResolver
}

type boundRoots struct {
	binder platformpostgres.OwnerBinder[*boundResolver]
	valid  bool
}

func bindRoots() (rootResolver, rootResolver, error) {
	binder, err := platformpostgres.NewOwnerBinder("analytics",
		func(database platformpostgres.BoundExecutor) *boundResolver {
			return &boundResolver{database: database}
		})
	if err != nil || queryRootSQL == "" || dashboardRootSQL == "" {
		return nil, nil, ErrResolve
	}
	bound := &boundRoots{binder: binder, valid: true}
	return bound.queryRoot, bound.dashboardRoot, nil
}

func (resolver *boundRoots) queryRoot(
	ctx context.Context, scope platformtx.TenantScope,
	releaseID, versionID uuid.UUID,
) (uuid.UUID, error) {
	return resolver.resolve(ctx, scope, queryRootSQL, releaseID, versionID)
}

func (resolver *boundRoots) dashboardRoot(
	ctx context.Context, scope platformtx.TenantScope,
	releaseID, versionID uuid.UUID,
) (uuid.UUID, error) {
	return resolver.resolve(ctx, scope, dashboardRootSQL, releaseID, versionID)
}

func (resolver *boundRoots) resolve(
	ctx context.Context, scope platformtx.TenantScope, statement string,
	releaseID, versionID uuid.UUID,
) (uuid.UUID, error) {
	if resolver == nil || !resolver.valid || statement == "" ||
		ctx == nil || ctx.Err() != nil || !validV7(releaseID) || !validV7(versionID) ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil ||
		platformtx.RequireOwner(scope, "analytics") != nil ||
		scope.OperationID() != "analytics.publication.execute" {
		return uuid.Nil, ErrResolve
	}
	bound, err := platformtx.Bind(scope, resolver.binder)
	if err != nil || bound == nil {
		return uuid.Nil, ErrResolve
	}
	var rootID uuid.UUID
	err = bound.database.QueryRow(ctx, statement, releaseID, versionID).Scan(&rootID)
	if err != nil || !validV7(rootID) {
		return uuid.Nil, ErrResolve
	}
	return rootID, nil
}

func (value *roots) QueryRoot(ctx context.Context, scope platformtx.TenantScope,
	releaseID, versionID uuid.UUID) (uuid.UUID, error) {
	return value.query(ctx, scope, releaseID, versionID)
}

func (value *roots) DashboardRoot(ctx context.Context, scope platformtx.TenantScope,
	releaseID, versionID uuid.UUID) (uuid.UUID, error) {
	return value.dashboard(ctx, scope, releaseID, versionID)
}

var _ Roots = (*roots)(nil)

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}
