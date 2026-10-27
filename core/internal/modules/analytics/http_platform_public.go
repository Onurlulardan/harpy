package analytics

import (
	"fmt"

	dashboardpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/dashboard"
	publicedgepostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/publicedge"
	publicexecutionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/publicexecution"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func buildPublicRuntime(
	dependencies PlatformHTTPRuntimeDependencies,
	values httpPlatformValues,
) (*publicationapp.PublicApplication, error) {
	roots, err := publicexecutionpostgres.NewResolver()
	if err != nil {
		return nil, fmt.Errorf("build Analytics public root resolver: %w", err)
	}
	queries, err := queryapp.NewPublicApplication(
		querypostgresBinder(), closedQueryContracts{}, closedQueryContracts{}, roots)
	if err != nil {
		return nil, fmt.Errorf("build Analytics public query application: %w", err)
	}
	executor, err := newPublicExecutor(
		queries, roots, dashboardpostgres.NewPublicSource())
	if err != nil {
		return nil, fmt.Errorf("build Analytics public executor: %w", err)
	}
	scopes, err := newPublicScopeCoordinator(
		dependencies.Generations, dependencies.Scopes)
	if err != nil {
		return nil, fmt.Errorf("build Analytics public scopes: %w", err)
	}
	directory, err := publicedgepostgres.NewDirectory(dependencies.PublicDB)
	if err != nil {
		return nil, fmt.Errorf("build Analytics public directory: %w", err)
	}
	limiter, err := publicedgepostgres.NewLimiter(dependencies.PublicDB, values)
	if err != nil {
		return nil, fmt.Errorf("build Analytics public limiter: %w", err)
	}
	application, err := publicationapp.NewPublicApplication(
		directory, scopes, executor, limiter, values)
	if err != nil {
		return nil, fmt.Errorf("build Analytics public application: %w", err)
	}
	return application, nil
}
