package billing

import (
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func NewTenantHTTPRuntime(
	resolver GenerationResolver,
	coordinator ScopeCoordinator,
	catalogs platformasync.WriteCatalogs,
) (billingapi.TenantRuntime, error) {
	clock := func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }
	return NewTenantOperations(resolver, coordinator, catalogs, uuid.NewV7, clock)
}
