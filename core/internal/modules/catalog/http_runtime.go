package catalog

import (
	"context"
	"fmt"
	"reflect"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	dependencyapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/dependency"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	registryapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/registry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type HTTPGenerationResolver interface {
	ResolveTenantAnalyticsGenerations(context.Context, uuid.UUID) (int64, int64, error)
	ResolveTenantBillingGenerations(context.Context, uuid.UUID) (int64, int64, int64, error)
	ResolveTenantCatalogGenerations(context.Context, uuid.UUID) (int64, int64, error)
	ResolveTenantDocumentsGenerations(context.Context, uuid.UUID) (int64, int64, error)
	ResolveTenantIAMGenerations(context.Context, uuid.UUID) (int64, int64, int64, error)
	ResolveTenantWorkflowGenerations(context.Context, uuid.UUID) (int64, int64, error)
}

type HTTPScopeCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type HTTPRuntime struct {
	resolver   HTTPGenerationResolver
	scopes     HTTPScopeCoordinator
	registry   *registryapp.Application
	dependency *dependencyapp.Application
	lifecycle  *lifecycleapp.Application
	menu       *menuapp.Application
}

func NewHTTPRuntime(resolver HTTPGenerationResolver, scopes HTTPScopeCoordinator,
	registry *registryapp.Application, dependency *dependencyapp.Application,
	lifecycle *lifecycleapp.Application,
	menu *menuapp.Application,
) (*HTTPRuntime, error) {
	if nilHTTPPort(resolver) || nilHTTPPort(scopes) || registry == nil || dependency == nil ||
		lifecycle == nil || menu == nil {
		return nil, fmt.Errorf("catalog HTTP runtime dependencies are required")
	}
	return &HTTPRuntime{resolver: resolver, scopes: scopes, registry: registry, dependency: dependency,
		lifecycle: lifecycle, menu: menu}, nil
}

var _ cataloghttp.Application = (*HTTPRuntime)(nil)

func nilHTTPPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}
