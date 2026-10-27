package analytics

import (
	"context"
	"fmt"
	"time"

	analyticscache "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/cache"
	querypostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/query"
	analyticsforms "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/formsreference"
	analyticsentitlement "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/entitlement"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PlatformQueryRuntimeDependencies struct {
	Generations  HTTPGenerationResolver
	Scopes       HTTPScopeCoordinator
	Entitlements billingapi.MeteredResourceMutationDecider
	Catalogs     platformasync.WriteCatalogs
}

func buildPlatformQueryApplication(
	dependencies PlatformQueryRuntimeDependencies,
	values httpPlatformValues,
	effects *AsyncEffects,
) (*queryapp.Application, error) {
	if nilHTTPPort(dependencies.Generations) || nilHTTPPort(dependencies.Scopes) ||
		nilHTTPPort(dependencies.Entitlements) || effects == nil {
		return nil, fmt.Errorf("Analytics query runtime dependencies are invalid")
	}
	cache, err := analyticscache.New(values, 4096)
	if err != nil {
		return nil, fmt.Errorf("build Analytics revision cache: %w", err)
	}
	contracts := closedQueryContracts{}
	queries, err := queryapp.NewApplication(querypostgres.NewBinder(), queryAuthorizer{},
		analyticsentitlement.MutationDecider(dependencies.Entitlements), contracts,
		contracts, cache, values, values, effects, contracts)
	if err != nil {
		return nil, fmt.Errorf("build Analytics query application: %w", err)
	}
	return queries, nil
}

type closedQueryContracts struct{}

func (closedQueryContracts) Dependencies(context.Context, platformtx.TenantScope,
	[]queryapp.DependencyRequest) ([]queryapp.DependencySnapshot, error) {
	return nil, errClosedQueryContracts
}

func (closedQueryContracts) Normalize(context.Context, platformtx.TenantScope,
	queryapp.Document, queryapp.Document, queryapp.Document, []queryapp.DependencySnapshot,
) (queryapp.NormalizedContract, error) {
	return queryapp.NormalizedContract{}, errClosedQueryContracts
}

func (closedQueryContracts) BindExecution(context.Context, platformtx.TenantScope,
	queryapp.Snapshot, []queryapp.DependencySnapshot, queryapp.Document,
) (queryapp.ExecutionBinding, error) {
	return queryapp.ExecutionBinding{}, errClosedQueryContracts
}

func (closedQueryContracts) Run(context.Context, platformtx.TenantScope, queryapp.Snapshot,
	[]queryapp.DependencySnapshot, queryapp.Document) (queryapp.ExecutionResult, error) {
	return queryapp.ExecutionResult{}, errClosedQueryContracts
}

func (closedQueryContracts) BindVersion(context.Context, platformtx.TenantScope,
	analyticsforms.Source, []analyticsforms.Target, time.Time) error {
	return errClosedQueryContracts
}

func (closedQueryContracts) DecommissionSource(context.Context, platformtx.TenantScope,
	analyticsforms.ID, *analyticsforms.ID, time.Time) error {
	return errClosedQueryContracts
}

var errClosedQueryContracts = fmt.Errorf("analytics forms queries are unavailable")
