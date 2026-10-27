package analyticsdashboardexport

import (
	"context"
	"errors"
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/jackc/pgx/v5"
)

func (store *Store) NextAnalyticsDashboardExportUnknown(
	ctx context.Context,
) (platformasync.AnalyticsDashboardExportUnknownCandidate, bool, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		return platformasync.AnalyticsDashboardExportUnknownCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportSourceUnavailable
	}
	candidate := platformasync.AnalyticsDashboardExportUnknownCandidate{}
	var requestDigest []byte
	err := store.database.QueryRow(ctx, unknownCandidateSQL).Scan(
		&candidate.TenantID, &candidate.TenantSchema, &candidate.ExportID,
		&candidate.JobID, &candidate.AttemptID, &candidate.AttemptGeneration,
		&candidate.LeaseOwnerID, &candidate.EffectID, &requestDigest,
		&candidate.UnknownAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return platformasync.AnalyticsDashboardExportUnknownCandidate{}, false, nil
	}
	if err != nil {
		return platformasync.AnalyticsDashboardExportUnknownCandidate{}, false,
			fmt.Errorf("read analytics dashboard export unknown candidate: %w", err)
	}
	if len(requestDigest) != len(candidate.EffectRequestSHA256) {
		return platformasync.AnalyticsDashboardExportUnknownCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportEvidenceInvalid
	}
	copy(candidate.EffectRequestSHA256[:], requestDigest)
	if !candidate.Valid() {
		return platformasync.AnalyticsDashboardExportUnknownCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportEvidenceInvalid
	}
	return candidate, true, nil
}
