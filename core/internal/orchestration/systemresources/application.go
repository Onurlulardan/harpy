package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type TenantCoordinator interface {
	InTenant(
		context.Context,
		platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error,
	) error
}

type Application struct {
	coordinator TenantCoordinator
	iam         publicresource.Port
	layout      layoutapi.ScopedResolver
}

func NewApplication(
	coordinator TenantCoordinator,
	iam publicresource.Port,
	layout layoutapi.ScopedResolver,
) (*Application, error) {
	if coordinator == nil || iam == nil || layout == nil {
		return nil, ErrContract
	}
	return &Application{coordinator: coordinator, iam: iam, layout: layout}, nil
}

func (application *Application) usable(ctx context.Context) bool {
	return application != nil && application.coordinator != nil &&
		application.iam != nil && application.layout != nil &&
		ctx != nil && ctx.Err() == nil
}

var _ Port = (*Application)(nil)
