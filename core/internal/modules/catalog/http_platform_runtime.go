package catalog

import (
	"fmt"

	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/lifecycle"
	menupostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/menu"
	registrypostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/registry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	dependencyapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/dependency"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	registryapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/registry"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type PlatformHTTPRuntimeDependencies struct {
	Generations HTTPGenerationResolver
	Scopes      HTTPScopeCoordinator
	Catalogs    platformasync.WriteCatalogs
	Keys        lifecycleapp.KeyResolver
	Owners      componentowner.Directory
	Demo        lifecycleapp.DemoOwner
}

func NewPlatformHTTPRuntime(dependencies PlatformHTTPRuntimeDependencies) (*HTTPRuntime, error) {
	if nilHTTPPort(dependencies.Generations) || nilHTTPPort(dependencies.Scopes) ||
		nilHTTPPort(dependencies.Catalogs) || nilHTTPPort(dependencies.Keys) ||
		nilHTTPPort(dependencies.Owners) || nilHTTPPort(dependencies.Demo) {
		return nil, fmt.Errorf("catalog platform HTTP dependencies are required")
	}
	values := runtimeValues{}
	async, err := NewAsyncPort(dependencies.Catalogs, values)
	if err != nil {
		return nil, fmt.Errorf("build catalog async port: %w", err)
	}
	registryBinder := registrypostgres.NewBinder()
	registry, err := registryapp.NewApplication(registryBinder, applicationAuthorizer{})
	if err != nil {
		return nil, fmt.Errorf("build catalog registry application: %w", err)
	}
	dependency, err := dependencyapp.NewApplication(registryBinder, applicationAuthorizer{})
	if err != nil {
		return nil, fmt.Errorf("build catalog dependency application: %w", err)
	}
	targets, err := menuapp.NewOwnerTargetAuthorizer(dependencies.Owners)
	if err != nil {
		return nil, err
	}
	menu, err := menuapp.NewApplication(menupostgres.NewBinder(), applicationAuthorizer{},
		targets, values, values)
	if err != nil {
		return nil, fmt.Errorf("build catalog menu application: %w", err)
	}
	lifecycle, err := lifecycleapp.NewApplication(lifecycleapp.Dependencies{
		Binder: lifecyclepostgres.NewBinder(), Registry: registryBinder, Keys: dependencies.Keys,
		Owners: dependencies.Owners, Authorizer: lifecycleAuthorizer{}, Jobs: async, Events: async,
		Backup: NewBackupVerifier(), Demo: dependencies.Demo,
		IDs: values, Clock: values,
	})
	if err != nil {
		return nil, fmt.Errorf("build catalog lifecycle application: %w", err)
	}
	return NewHTTPRuntime(dependencies.Generations, dependencies.Scopes, registry, dependency,
		lifecycle, menu)
}
