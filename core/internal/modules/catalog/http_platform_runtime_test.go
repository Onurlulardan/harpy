package catalog

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type catalogHTTPScopesFixture struct{ HTTPScopeCoordinator }
type catalogHTTPAccessFixture struct{ HTTPAccessVerifier }

func TestCatalogPlatformHTTPRuntimeRequiresDemoOwner(t *testing.T) {
	dependencies := catalogPlatformHTTPDependencies(t)
	dependencies.Demo = nil
	if runtime, err := NewPlatformHTTPRuntime(dependencies); err == nil || runtime != nil {
		t.Fatal("Catalog platform HTTP runtime accepted a nil demo owner")
	}
	var typedNil *catalogDemoOwnerFixture
	dependencies.Demo = typedNil
	if runtime, err := NewPlatformHTTPRuntime(dependencies); err == nil || runtime != nil {
		t.Fatal("Catalog platform HTTP runtime accepted a typed-nil demo owner")
	}
}

func TestCatalogPlatformHTTPRuntimeAcceptsCompleteDependencies(t *testing.T) {
	if runtime, err := NewPlatformHTTPRuntime(catalogPlatformHTTPDependencies(t)); err != nil ||
		runtime == nil {
		t.Fatalf("build Catalog platform HTTP runtime: runtime=%v err=%v", runtime, err)
	}
}

func TestCatalogPlatformHTTPHandlerPublishesExactUniqueRoutes(t *testing.T) {
	runtime, err := NewPlatformHTTPRuntime(catalogPlatformHTTPDependencies(t))
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHTTPHandler(HTTPHandlerDependencies{
		Access: &catalogHTTPAccessFixture{}, Application: runtime, Report: func(error) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	routes := handler.Routes()
	if len(routes) != 32 {
		t.Fatalf("Catalog route count = %d, want 32", len(routes))
	}
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		if _, found := seen[key]; found {
			t.Fatalf("duplicate Catalog route %s", key)
		}
		seen[key] = struct{}{}
	}
	handler.routes[1] = handler.routes[0]
	if handler.Ready() || handler.Routes() != nil {
		t.Fatal("Catalog handler accepted a duplicate route manifest")
	}
}

func catalogPlatformHTTPDependencies(t *testing.T) PlatformHTTPRuntimeDependencies {
	t.Helper()
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	owner, err := NewCatalogComponentOwner()
	if err != nil {
		t.Fatal(err)
	}
	owners, err := NewComponentOwnerDirectory(owner, owner, owner)
	if err != nil {
		t.Fatal(err)
	}
	return PlatformHTTPRuntimeDependencies{
		Generations: &workerGenerationsFixture{}, Scopes: &catalogHTTPScopesFixture{},
		Catalogs: catalogs, Keys: &workerKeysFixture{}, Owners: owners,
		Demo: &catalogDemoOwnerFixture{},
	}
}
