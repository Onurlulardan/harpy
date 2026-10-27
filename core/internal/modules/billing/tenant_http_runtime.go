package billing

import (
	"fmt"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func NewTenantOperations(
	resolver billingapi.TenantGenerationResolver,
	coordinator billingapi.TenantScopeCoordinator,
	catalogs platformasync.WriteCatalogs,
	nextID func() (uuid.UUID, error),
	clock func() time.Time,
) (billingapi.TenantRuntime, error) {
	if nextID == nil || clock == nil {
		return nil, fmt.Errorf("tenant Billing operation identity dependencies are required")
	}
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, fmt.Errorf("tenant Billing async catalogs are invalid: %w", err)
	}
	bindings := &tenantHTTPBindings{
		catalogs: catalogs, nextID: nextID, clock: clock,
	}
	return newTenantOperations(resolver, coordinator, bindings)
}
