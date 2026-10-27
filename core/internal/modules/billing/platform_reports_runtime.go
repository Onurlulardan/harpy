package billing

import (
	"fmt"

	moduleapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type PlatformReportsDependencies struct {
	GlobalCatalog *moduleapp.GlobalCatalog
	Queries       PlatformReportQueries
	Locator       PlatformTransitionResolver
	Coordinator   ScopeCoordinator
	Bindings      Bindings
}

type platformReports struct {
	global      *moduleapp.GlobalCatalog
	queries     PlatformReportQueries
	locator     PlatformTransitionResolver
	coordinator ScopeCoordinator
	bindings    Bindings
}

func NewPlatformReports(
	dependencies PlatformReportsDependencies,
) (billingapi.PlatformReports, error) {
	if dependencies.GlobalCatalog == nil || dependencies.Queries == nil ||
		dependencies.Locator == nil || dependencies.Coordinator == nil ||
		dependencies.Bindings == nil {
		return nil, fmt.Errorf("billing platform reports dependencies are required")
	}
	return &platformReports{
		global: dependencies.GlobalCatalog, queries: dependencies.Queries,
		locator: dependencies.Locator, coordinator: dependencies.Coordinator,
		bindings: dependencies.Bindings,
	}, nil
}

var _ billingapi.PlatformReports = (*platformReports)(nil)
