package worker

import (
	"fmt"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	catalogmodule "github.com/Onurlulardan/harpy/core/internal/modules/catalog"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildCatalogJobProcessor(
	settings config.Worker,
	release platformasync.ReleaseCatalogs,
	catalogs platformasync.WriteCatalogs,
	pool *postgres.Pool,
	coordinator *platformtx.Coordinator,
) (jobexecution.Processor, error) {
	if pool == nil || coordinator == nil {
		return nil, fmt.Errorf("Catalog worker composition dependencies are required")
	}
	owners, err := buildCatalogWorkerOwners()
	if err != nil {
		return nil, err
	}
	keys, err := settings.CatalogPackageKeyring()
	if err != nil {
		return nil, fmt.Errorf("load Catalog worker package keyring: %w", err)
	}
	processor, err := catalogmodule.NewPlatformJobProcessor(
		catalogmodule.PlatformWorkerDependencies{
			Release: release, Catalogs: catalogs, Generations: pool,
			Scopes: coordinator, Keys: keys, Owners: owners,
			Demo: catalogmodule.NewClosedDemoOwner(),
		})
	if err != nil {
		return nil, fmt.Errorf("build Catalog operation processor: %w", err)
	}
	return processor, nil
}

func buildCatalogWorkerOwners() (catalogmodule.ComponentOwnerDirectory, error) {
	analytics, err := analyticsmodule.NewCatalogComponentOwner()
	if err != nil {
		return nil, fmt.Errorf("build Analytics Catalog owner: %w", err)
	}
	catalogOwner, err := catalogmodule.NewCatalogComponentOwner()
	if err != nil {
		return nil, fmt.Errorf("build Catalog component owner: %w", err)
	}
	iam, err := iammodule.NewCatalogComponentOwner()
	if err != nil {
		return nil, fmt.Errorf("build IAM Catalog owner: %w", err)
	}
	directory, err := catalogmodule.NewComponentOwnerDirectory(
		analytics, catalogOwner, iam,
	)
	if err != nil {
		return nil, err
	}
	return directory, nil
}
