package documents

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/clamav"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type configuredProductionResources struct {
	artifacts platformArtifacts
	billing   platformBilling
	jobs      *APIAsyncPort
	effects   uploadapp.EffectRunner
}

func openConfiguredProductionResources(ctx context.Context,
	artifacts config.DocumentsArtifacts, clamAVAddress string,
	catalogs platformasync.WriteCatalogs, billing ConfiguredDocumentsBilling,
) (configuredProductionResources, error) {
	if ctx == nil || ctx.Err() != nil || artifacts.Validate() != nil ||
		clamAVAddress == "" || nilHTTPPort(catalogs) || nilHTTPPort(billing) {
		return configuredProductionResources{},
			fmt.Errorf("Documents configured production dependencies are required")
	}
	opened, err := openPlatformArtifacts(ctx, artifacts)
	if err != nil {
		return configuredProductionResources{},
			fmt.Errorf("open Documents production artifacts: %w", err)
	}
	return configureProductionResources(ctx, opened, clamAVAddress, catalogs, billing)
}

func configureProductionResources(ctx context.Context, artifacts platformArtifacts,
	clamAVAddress string, catalogs platformasync.WriteCatalogs,
	billing ConfiguredDocumentsBilling,
) (configuredProductionResources, error) {
	if ctx == nil || ctx.Err() != nil || artifacts.objects == nil ||
		artifacts.keyring == nil || artifacts.cipher == nil || clamAVAddress == "" ||
		nilHTTPPort(catalogs) || nilHTTPPort(billing) {
		return configuredProductionResources{},
			fmt.Errorf("Documents configured production resources are required")
	}
	scanner, err := clamav.NewVerified(ctx, clamAVAddress, &net.Dialer{
		Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}, platformClock{})
	if err != nil {
		return configuredProductionResources{}, fmt.Errorf("verify Documents ClamAV: %w", err)
	}
	effects, err := uploadapp.NewEffectPipeline(artifacts.keyring, artifacts.cipher,
		scanner, artifacts.objects, platformClock{})
	if err != nil {
		return configuredProductionResources{},
			fmt.Errorf("build Documents upload effects: %w", err)
	}
	usage, err := newPlatformBilling(billing)
	if err != nil {
		return configuredProductionResources{}, err
	}
	jobs, err := NewAPIAsyncPort(catalogs, platformClock{})
	if err != nil {
		return configuredProductionResources{},
			fmt.Errorf("build Documents async API port: %w", err)
	}
	return configuredProductionResources{artifacts: artifacts, billing: usage,
		jobs: jobs, effects: effects}, nil
}
