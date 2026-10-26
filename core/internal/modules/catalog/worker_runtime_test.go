package catalog

import (
	"testing"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type workerGenerationsFixture struct{ HTTPGenerationResolver }
type workerScopesFixture struct{ WorkerScopeCoordinator }
type workerKeysFixture struct{ lifecycleapp.KeyResolver }
type catalogDemoOwnerFixture struct{ lifecycleapp.DemoOwner }

func TestCatalogPlatformJobProcessorPublishesExactResumePolicy(t *testing.T) {
	release, err := platformasync.LoadReleaseCatalogs()
	if err != nil {
		t.Fatal(err)
	}
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
	processor, err := NewPlatformJobProcessor(PlatformWorkerDependencies{
		Release: release, Catalogs: catalogs,
		Generations: &workerGenerationsFixture{}, Scopes: &workerScopesFixture{},
		Keys: &workerKeysFixture{}, Owners: owners, Demo: &catalogDemoOwnerFixture{},
	})
	if err != nil {
		t.Fatal(err)
	}
	policy := processor.Policy()
	if policy.Kind != asyncmanifestv1.CatalogOperationResumeJobKind ||
		policy.PayloadRevision != 1 || policy.OwnerDomain != "catalog" {
		t.Fatalf("Catalog worker policy drifted: %#v", policy)
	}
}

func TestCatalogPlatformJobProcessorRejectsIncompleteDependencies(t *testing.T) {
	if processor, err := NewPlatformJobProcessor(PlatformWorkerDependencies{}); err == nil ||
		processor != nil {
		t.Fatal("incomplete Catalog worker dependencies accepted")
	}
	dependencies := catalogPlatformWorkerDependencies(t)
	dependencies.Demo = nil
	if processor, err := NewPlatformJobProcessor(dependencies); err == nil || processor != nil {
		t.Fatal("Catalog worker accepted a nil demo owner")
	}
	var typedNil *catalogDemoOwnerFixture
	dependencies.Demo = typedNil
	if processor, err := NewPlatformJobProcessor(dependencies); err == nil || processor != nil {
		t.Fatal("Catalog worker accepted a typed-nil demo owner")
	}
}

func catalogPlatformWorkerDependencies(t *testing.T) PlatformWorkerDependencies {
	t.Helper()
	release, err := platformasync.LoadReleaseCatalogs()
	if err != nil {
		t.Fatal(err)
	}
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
	return PlatformWorkerDependencies{
		Release: release, Catalogs: catalogs,
		Generations: &workerGenerationsFixture{}, Scopes: &workerScopesFixture{},
		Keys: &workerKeysFixture{}, Owners: owners, Demo: &catalogDemoOwnerFixture{},
	}
}
