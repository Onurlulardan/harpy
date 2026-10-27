package policypostgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantScope = transaction.TenantScope

type boundVersions struct {
	database platformpostgres.BoundExecutor
}

type ports struct {
	versionBinder platformpostgres.OwnerBinder[boundVersions]
}

func newPorts() (*ports, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		catalog.Owner,
		func(database platformpostgres.BoundExecutor) boundVersions {
			return boundVersions{database: database}
		},
	)
	if err != nil {
		return nil, catalog.ErrContract
	}
	return &ports{versionBinder: binder}, nil
}

func (ports *ports) bindVersions(scope tenantScope) (boundVersions, error) {
	if ports == nil {
		return boundVersions{}, catalog.ErrContract
	}
	bound, err := transaction.Bind(scope, ports.versionBinder)
	if err != nil {
		return boundVersions{}, catalog.ErrContract
	}
	return bound, nil
}

func scopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) &&
		transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.SnapshotQuery) == nil
}
