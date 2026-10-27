package dashboardexport

import (
	"context"
	"regexp"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

var workerTenantSchemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

func (storage *WorkerStore) PeekRenderCandidate(
	ctx context.Context,
) (exportapp.RenderCandidate, bool, error) {
	if storage == nil || storage.source == nil ||
		ctx == nil || ctx.Err() != nil {
		return exportapp.RenderCandidate{}, false, exportapp.ErrUnavailable
	}
	stored, found, err := storage.source.NextAnalyticsDashboardExport(ctx)
	if err != nil {
		return exportapp.RenderCandidate{}, false, exportapp.ErrUnavailable
	}
	if !found {
		return exportapp.RenderCandidate{}, false, nil
	}
	candidate := exportapp.RenderCandidate{TenantID: stored.TenantID,
		TenantSchema: stored.TenantSchema, ExportID: stored.ExportID,
		JobID: stored.JobID, AttemptGeneration: stored.AttemptGeneration}
	if !validWorkerCandidate(candidate) {
		return exportapp.RenderCandidate{}, false, exportapp.ErrUnavailable
	}
	return candidate, true, nil
}

func validWorkerCandidate(candidate exportapp.RenderCandidate) bool {
	return candidate.TenantID != [16]byte{} && candidate.ExportID.Version() == 7 &&
		candidate.JobID.Version() == 7 &&
		workerTenantSchemaPattern.MatchString(candidate.TenantSchema) &&
		candidate.AttemptGeneration >= 1 &&
		candidate.AttemptGeneration <= exportapp.MaximumAttempts
}
