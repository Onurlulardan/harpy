package api

import (
	"fmt"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	catalogmodule "github.com/Onurlulardan/harpy/core/internal/modules/catalog"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type catalogHTTPRuntime struct {
	handler     *catalogmodule.HTTPHandler
	application *catalogmodule.HTTPRuntime
}

func (runtime catalogHTTPRuntime) Ready() bool {
	return runtime.handler != nil && runtime.handler.Ready() && runtime.application != nil
}

func (runtime catalogHTTPRuntime) Routes() []httpedge.Route {
	if !runtime.Ready() {
		return nil
	}
	return runtime.handler.Routes()
}

func buildCatalogHTTP(
	settings config.API,
	pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs,
	iam iamAPIRuntime,
	report httpedge.ErrorSink,
) (catalogHTTPRuntime, error) {
	if pool == nil || iam.catalogAccess == nil || report == nil {
		return catalogHTTPRuntime{}, fmt.Errorf("Catalog HTTP composition dependencies are required")
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return catalogHTTPRuntime{}, fmt.Errorf("build Catalog scope coordinator: %w", err)
	}
	access, err := newCatalogAccessVerifier(iam.catalogAccess)
	if err != nil {
		return catalogHTTPRuntime{}, err
	}
	owners, err := buildCatalogOwners()
	if err != nil {
		return catalogHTTPRuntime{}, err
	}
	keys, err := settings.CatalogPackageKeyring()
	if err != nil {
		return catalogHTTPRuntime{}, fmt.Errorf("load Catalog package keyring: %w", err)
	}
	application, err := catalogmodule.NewPlatformHTTPRuntime(
		catalogmodule.PlatformHTTPRuntimeDependencies{
			Generations: pool, Scopes: coordinator, Catalogs: catalogs,
			Keys: keys, Owners: owners, Demo: catalogmodule.NewClosedDemoOwner(),
		})
	if err != nil {
		return catalogHTTPRuntime{}, fmt.Errorf("build Catalog HTTP application: %w", err)
	}
	handler, err := catalogmodule.NewHTTPHandler(catalogmodule.HTTPHandlerDependencies{
		Access: access, Application: application, Report: catalogmodule.HTTPErrorSink(report),
	})
	if err != nil {
		return catalogHTTPRuntime{}, fmt.Errorf("build Catalog HTTP handler: %w", err)
	}
	if !handler.Ready() {
		return catalogHTTPRuntime{}, fmt.Errorf("build Catalog HTTP handler: not ready")
	}
	return catalogHTTPRuntime{handler: handler, application: application}, nil
}

func buildCatalogOwners() (catalogmodule.ComponentOwnerDirectory, error) {
	analytics, err := analyticsmodule.NewCatalogComponentOwner()
	if err != nil {
		return nil, fmt.Errorf("build Analytics Catalog component owner: %w", err)
	}
	catalogOwner, err := catalogmodule.NewCatalogComponentOwner()
	if err != nil {
		return nil, fmt.Errorf("build Catalog component owner: %w", err)
	}
	iam, err := iammodule.NewCatalogComponentOwner()
	if err != nil {
		return nil, fmt.Errorf("build IAM Catalog component owner: %w", err)
	}
	directory, err := catalogmodule.NewComponentOwnerDirectory(
		analytics, catalogOwner, iam,
	)
	if err != nil {
		return nil, err
	}
	return directory, nil
}
