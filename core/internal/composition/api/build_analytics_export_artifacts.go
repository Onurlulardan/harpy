package api

import (
	"context"
	"fmt"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject"
	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func buildAnalyticsExportContent(
	ctx context.Context, settings config.API,
) (analyticsmodule.DashboardExportContentRuntime, error) {
	artifacts, err := settings.AnalyticsDashboardExportArtifacts()
	if err != nil {
		return nil, fmt.Errorf("load Analytics export artifacts: %w", err)
	}
	storage, err := artifacts.Storage()
	if err != nil {
		return nil, fmt.Errorf("load Analytics export storage: %w", err)
	}
	objects, err := artifactobject.OpenS3(ctx, storage,
		analyticsmodule.DashboardExportObjectPrefix,
		analyticsmodule.DashboardExportObjectContract,
		analyticsmodule.DashboardExportMaximumCiphertextBytes)
	if err != nil {
		return nil, fmt.Errorf("open Analytics export object store: %w", err)
	}
	keys, err := analyticsmodule.NewDashboardExportArtifactKeyring(artifacts)
	if err != nil {
		return nil, err
	}
	runtime, err := analyticsmodule.NewDashboardExportContentRuntime(
		objects, keys, analyticsmodule.DashboardExportObjectPrefix,
	)
	if err != nil {
		return nil, fmt.Errorf("build Analytics export content runtime: %w", err)
	}
	return runtime, nil
}
