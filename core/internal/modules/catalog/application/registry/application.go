package registry

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const permissionRead = "catalog.read"

type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, string) error
}

type Application struct {
	binder     packageregistry.Binder
	authorizer Authorizer
}

func NewApplication(
	binder packageregistry.Binder, authorizer Authorizer,
) (*Application, error) {
	if binder == nil || authorizer == nil {
		return nil, fmt.Errorf("Catalog registry dependencies are required")
	}
	return &Application{binder: binder, authorizer: authorizer}, nil
}

func (application *Application) bind(
	ctx context.Context, scope platformtx.TenantScope, operation string,
) (packageregistry.Reader, error) {
	if application == nil || application.binder == nil || application.authorizer == nil ||
		ctx == nil || ctx.Err() != nil {
		return nil, packageregistry.ErrUnavailable
	}
	if !platformtx.ContextOwnsScope(ctx, scope) || scope.Mode() != platformtx.SnapshotQuery ||
		scope.OperationID() != operation {
		return nil, packageregistry.ErrForbidden
	}
	if err := application.authorizer.Authorize(ctx, scope, permissionRead); err != nil {
		return nil, packageregistry.ErrForbidden
	}
	registry, err := application.binder.Bind(ctx, scope)
	if err != nil || registry == nil {
		return nil, packageregistry.ErrUnavailable
	}
	return registry, nil
}
