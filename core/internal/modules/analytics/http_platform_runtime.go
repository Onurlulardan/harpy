package analytics

import (
	"fmt"

	dashboardpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/dashboard"
	exportpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/dashboardexport"
	publicationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/publication"
	publicedgepostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/publicedge"
	resourcepostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/resource"
	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	analyticsentitlement "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/entitlement"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type PlatformHTTPRuntimeDependencies struct {
	Generations     HTTPGenerationResolver
	Scopes          HTTPScopeCoordinator
	PublicDB        publicedgepostgres.Database
	Legacy          compatibilityapp.PublicLegacyStore
	Principals      dashboardPrincipalAuthorizer
	Entitlements    analyticsentitlement.MutationDecider
	Catalogs        platformasync.WriteCatalogs
	ExportScheduler exportapp.Scheduler
	ExportContent   exportapp.ContentRuntime
}

func NewPlatformHTTPRuntime(
	dependencies PlatformHTTPRuntimeDependencies,
) (*HTTPRuntime, *publicationapp.PublicApplication, *compatibilityapp.Application, error) {
	if nilHTTPPort(dependencies.Generations) || nilHTTPPort(dependencies.Scopes) ||
		nilHTTPPort(dependencies.PublicDB) || nilHTTPPort(dependencies.Legacy) ||
		nilHTTPPort(dependencies.Principals) || nilHTTPPort(dependencies.Entitlements) {
		// Owner ports are final-composition dependencies; absence is fail-closed.
		return nil, nil, nil, fmt.Errorf("Analytics platform HTTP dependencies are invalid")
	}
	if nilHTTPPort(dependencies.ExportScheduler) || nilHTTPPort(dependencies.ExportContent) {
		return nil, nil, nil, fmt.Errorf("Analytics dashboard export dependencies are invalid")
	}
	values := httpPlatformValues{}
	effects, err := NewAsyncEffects(dependencies.Catalogs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Analytics async effects: %w", err)
	}
	queries, err := buildPlatformQueryApplication(PlatformQueryRuntimeDependencies{
		Generations: dependencies.Generations, Scopes: dependencies.Scopes,
		Entitlements: dependencies.Entitlements, Catalogs: dependencies.Catalogs,
	}, values, effects)
	if err != nil {
		return nil, nil, nil, err
	}
	resources := resourcepostgres.NewVerifier()
	dashboards, err := dashboardapp.NewApplication(dashboardpostgres.NewBinder(),
		dashboardAuthorizer{principals: dependencies.Principals}, dependencies.Entitlements,
		resources, values, values, effects)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Analytics dashboard application: %w", err)
	}
	publications, err := publicationapp.NewApplication(publicationpostgres.NewBinder(),
		publicationAuthorizer{principals: dependencies.Principals}, dependencies.Entitlements, resources,
		values, values, values, effects)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Analytics publication application: %w", err)
	}
	exports, err := exportapp.NewApplication(exportpostgres.NewBinder(),
		dashboardExportAuthorizer{}, values, values,
		dependencies.ExportScheduler, dependencies.ExportContent)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Analytics dashboard export application: %w", err)
	}
	runtime, err := NewHTTPRuntime(dependencies.Generations, dependencies.Scopes,
		queries, dashboards, publications, exports)
	if err != nil {
		return nil, nil, nil, err
	}
	public, err := buildPlatformPublicRuntime(dependencies, values)
	if err != nil {
		return nil, nil, nil, err
	}
	legacy, err := compatibilityapp.NewLegacyApplication(
		dependencies.Legacy, compatibilityClock{},
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Analytics legacy application: %w", err)
	}
	return runtime, public, legacy, nil
}

func buildPlatformPublicRuntime(
	dependencies PlatformHTTPRuntimeDependencies,
	values httpPlatformValues,
) (*publicationapp.PublicApplication, error) {
	return buildPublicRuntime(dependencies, values)
}
