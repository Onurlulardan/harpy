package billing

import (
	"fmt"
	"time"

	platformreportspostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/platformreports"
	executionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/restquery/execution"
	moduleapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformreportapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformreports"
	billingreports "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/reports"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func NewPlatformReportsRuntime(
	coordinator *platformtx.Coordinator,
	keyring config.CursorKeyring,
	catalogs platformasync.WriteCatalogs,
) (billingapi.PlatformReports, error) {
	if coordinator == nil {
		return nil, fmt.Errorf("billing platform report coordinator is required")
	}
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, fmt.Errorf("billing platform report async catalogs are invalid: %w", err)
	}
	global, err := moduleapp.NewGlobalCatalog(coordinator)
	if err != nil {
		return nil, err
	}
	keys, err := globalHTTPQueryKeys(keyring)
	if err != nil {
		return nil, err
	}
	owner, err := executionpostgres.New()
	if err != nil {
		return nil, err
	}
	queries, err := billingreports.NewService(owner, keys)
	if err != nil {
		return nil, err
	}
	resolver, err := platformreportapp.NewResolver(platformreportspostgres.NewBinder())
	if err != nil {
		return nil, err
	}
	clock := func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }
	bindings := &tenantHTTPBindings{catalogs: catalogs, nextID: uuid.NewV7, clock: clock}
	return NewPlatformReports(PlatformReportsDependencies{
		GlobalCatalog: global, Queries: queries, Locator: resolver,
		Coordinator: coordinator, Bindings: bindings,
	})
}
