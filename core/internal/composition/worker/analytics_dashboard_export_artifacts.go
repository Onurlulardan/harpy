package worker

import (
	"context"
	"fmt"
	"path/filepath"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool"
	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func openAnalyticsDashboardExportArtifacts(
	ctx context.Context, settings config.Worker,
) (*analyticsmodule.DashboardExportArtifactRuntime, error) {
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
	directory := settings.AnalyticsDashboardExportSpoolDirectory()
	spools, err := artifactspool.NewManager(filepath.Join(directory.Value(), "runtime"))
	if err != nil {
		return nil, fmt.Errorf("open Analytics export spool: %w", err)
	}
	runtime, err := analyticsmodule.NewDashboardExportArtifactRuntime(
		spools, objects, keys, analyticsmodule.DashboardExportObjectPrefix,
	)
	if err != nil {
		return nil, fmt.Errorf("build Analytics export artifacts: %w", err)
	}
	return runtime, nil
}
