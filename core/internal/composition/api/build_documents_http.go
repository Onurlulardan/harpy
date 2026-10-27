package api

import (
	"context"
	"fmt"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	documentsmodule "github.com/Onurlulardan/harpy/core/internal/modules/documents"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildDocumentsHTTP(ctx context.Context, settings config.API,
	pool *postgres.Pool, catalogs platformasync.WriteCatalogs, iam iamAPIRuntime,
	report httpedge.ErrorSink, info buildinfo.Info,
) (*documentsmodule.HTTPHandler, *documentsmodule.Production, error) {
	if ctx == nil || ctx.Err() != nil || pool == nil ||
		iam.documentsAccess == nil || report == nil {
		return nil, nil, fmt.Errorf("Documents HTTP composition dependencies are required")
	}
	configured, err := settings.Documents()
	if err != nil {
		return nil, nil, fmt.Errorf("load Documents API settings: %w", err)
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return nil, nil, fmt.Errorf("build Documents scope coordinator: %w", err)
	}
	artifactSettings, err := configured.Artifacts()
	if err != nil {
		return nil, nil, fmt.Errorf("load Documents artifact settings: %w", err)
	}
	billingActorID, err := artifactSettings.BillingActorID()
	if err != nil {
		return nil, nil, fmt.Errorf("load Documents Billing actor: %w", err)
	}
	billing, err := billingmodule.NewConfiguredDocumentsUsage(catalogs, billingActorID)
	if err != nil {
		return nil, nil, fmt.Errorf("build Documents Billing usage: %w", err)
	}
	production, err := documentsmodule.NewConfiguredAPIProduction(ctx,
		documentsmodule.ConfiguredPlatformHTTPRuntimeDependencies{
			Settings: configured, Catalogs: catalogs, Generations: pool,
			Scopes: coordinator, Billing: billing, BuildID: info.Version().Commit,
		})
	if err != nil {
		return nil, nil, fmt.Errorf("build Documents API production: %w", err)
	}
	access, err := newDocumentsAccessVerifier(iam.documentsAccess)
	if err != nil {
		return nil, nil, err
	}
	handler, err := documentsmodule.NewHTTPHandler(documentsmodule.HTTPHandlerDependencies{
		Access: access, Application: production.Application(),
		Report: documentsmodule.HTTPErrorSink(report),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build Documents HTTP handler: %w", err)
	}
	if !handler.Ready() {
		return nil, nil, fmt.Errorf("Documents HTTP handler is not ready")
	}
	return handler, production, nil
}
