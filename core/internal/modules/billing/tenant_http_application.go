package billing

import (
	"fmt"
	"reflect"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type tenantOperations struct {
	resolver    billingapi.TenantGenerationResolver
	coordinator billingapi.TenantScopeCoordinator
	bindings    billingapi.TenantOperationBindings
}

func newTenantOperations(
	resolver billingapi.TenantGenerationResolver,
	coordinator billingapi.TenantScopeCoordinator,
	bindings billingapi.TenantOperationBindings,
) (*tenantOperations, error) {
	if nilDependency(resolver) || nilDependency(coordinator) || nilDependency(bindings) {
		return nil, fmt.Errorf("tenant Billing HTTP application dependencies are required")
	}
	return &tenantOperations{
		resolver: resolver, coordinator: coordinator, bindings: bindings,
	}, nil
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
