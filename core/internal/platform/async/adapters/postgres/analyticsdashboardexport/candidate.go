package analyticsdashboardexport

import (
	"context"
	"errors"
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/jackc/pgx/v5"
)

func (store *Store) NextAnalyticsDashboardExport(
	ctx context.Context,
) (platformasync.AnalyticsDashboardExportCandidate, bool, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		return platformasync.AnalyticsDashboardExportCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportSourceUnavailable
	}
	candidate := platformasync.AnalyticsDashboardExportCandidate{}
	err := store.database.QueryRow(ctx, candidateSQL).Scan(
		&candidate.TenantID, &candidate.TenantSchema, &candidate.ExportID,
		&candidate.JobID, &candidate.AttemptGeneration,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return platformasync.AnalyticsDashboardExportCandidate{}, false, nil
	}
	if err != nil {
		return platformasync.AnalyticsDashboardExportCandidate{}, false,
			fmt.Errorf("read analytics dashboard export candidate: %w", err)
	}
	if !candidate.Valid() {
		return platformasync.AnalyticsDashboardExportCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportEvidenceInvalid
	}
	return candidate, true, nil
}
