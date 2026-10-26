package documents

import (
	"context"
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type ConfiguredProductionDependencies struct {
	Artifacts     config.DocumentsArtifacts
	ClamAVAddress string
	Catalogs      platformasync.WriteCatalogs
	Generations   HTTPGenerationResolver
	Scopes        HTTPScopeCoordinator
	Billing       ConfiguredDocumentsBilling
	BuildID       string
}

type ConfiguredPlatformHTTPRuntimeDependencies struct {
	Settings    config.DocumentsAPI
	Catalogs    platformasync.WriteCatalogs
	Generations HTTPGenerationResolver
	Scopes      HTTPScopeCoordinator
	Billing     ConfiguredDocumentsBilling
	BuildID     string
}

func NewConfiguredProduction(ctx context.Context,
	dependencies ConfiguredProductionDependencies,
) (*Production, error) {
	if ctx == nil || ctx.Err() != nil || nilHTTPPort(dependencies.Generations) ||
		nilHTTPPort(dependencies.Scopes) || dependencies.BuildID == "" ||
		len(dependencies.BuildID) > 128 {
		return nil, fmt.Errorf("Documents configured production dependencies are required")
	}
	resources, err := openConfiguredProductionResources(ctx, dependencies.Artifacts,
		dependencies.ClamAVAddress, dependencies.Catalogs, dependencies.Billing)
	if err != nil {
		return nil, err
	}
	return buildProduction(resources, dependencies.Generations,
		dependencies.Scopes, dependencies.BuildID)
}

func NewConfiguredAPIProduction(ctx context.Context,
	dependencies ConfiguredPlatformHTTPRuntimeDependencies,
) (*Production, error) {
	if ctx == nil || ctx.Err() != nil || nilHTTPPort(dependencies.Billing) ||
		nilHTTPPort(dependencies.Generations) || nilHTTPPort(dependencies.Scopes) {
		return nil, fmt.Errorf("Documents configured API dependencies are required")
	}
	artifacts, err := dependencies.Settings.Artifacts()
	if err != nil {
		return nil, fmt.Errorf("load Documents artifact settings: %w", err)
	}
	address, err := dependencies.Settings.ClamAVAddress()
	if err != nil {
		return nil, err
	}
	return NewConfiguredProduction(ctx, ConfiguredProductionDependencies{
		Artifacts: artifacts, ClamAVAddress: address, Catalogs: dependencies.Catalogs,
		Generations: dependencies.Generations, Scopes: dependencies.Scopes,
		Billing: dependencies.Billing, BuildID: dependencies.BuildID,
	})
}

func NewConfiguredPlatformHTTPRuntime(ctx context.Context,
	dependencies ConfiguredPlatformHTTPRuntimeDependencies,
) (*HTTPRuntime, error) {
	production, err := NewConfiguredAPIProduction(ctx, dependencies)
	if err != nil {
		return nil, err
	}
	runtime, ok := production.application.(*ApplicationRuntime)
	if !ok || runtime == nil {
		return nil, fmt.Errorf("Documents configured application runtime is unavailable")
	}
	return runtime, nil
}
