package worker

import (
	"context"
	"fmt"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	documentsmodule "github.com/Onurlulardan/harpy/core/internal/modules/documents"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildDocumentsJobProcessors(ctx context.Context, settings config.Worker,
	release platformasync.ReleaseCatalogs, catalogs platformasync.WriteCatalogs,
	store asyncpostgres.WorkerStore, pool *postgres.Pool,
	coordinator *platformtx.Coordinator, buildID string,
) (*documentsmodule.WorkerProduction, error) {
	configured, err := settings.Documents()
	if err != nil {
		return nil, fmt.Errorf("load Documents worker settings: %w", err)
	}
	access, err := iammodule.NewDocumentsArchiveAccessRuntime(pool, coordinator)
	if err != nil {
		return nil, fmt.Errorf("build Documents archive access: %w", err)
	}
	artifactSettings, err := configured.Artifacts()
	if err != nil {
		return nil, fmt.Errorf("load Documents artifact settings: %w", err)
	}
	billingActorID, err := artifactSettings.BillingActorID()
	if err != nil {
		return nil, fmt.Errorf("load Documents Billing actor: %w", err)
	}
	billing, err := billingmodule.NewConfiguredDocumentsUsage(catalogs, billingActorID)
	if err != nil {
		return nil, fmt.Errorf("build Documents Billing usage: %w", err)
	}
	clamAVAddress, err := configured.ClamAVAddress()
	if err != nil {
		return nil, fmt.Errorf("load Documents ClamAV address: %w", err)
	}
	production, err := documentsmodule.NewConfiguredWorkerProduction(ctx,
		documentsmodule.ConfiguredWorkerProductionDependencies{
			Jobs: documentsmodule.ConfiguredPlatformJobDependencies{
				Settings: configured, Release: release, Catalogs: catalogs,
				Store: store, Scopes: coordinator, ArchiveAccess: access, Billing: billing,
			},
			ClamAVAddress: clamAVAddress, Generations: pool,
			ApplicationScopes: coordinator, BuildID: buildID,
		})
	if err != nil {
		return nil, fmt.Errorf("build Documents worker production: %w", err)
	}
	return production, nil
}
