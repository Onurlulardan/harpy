package analyticsdashboardexport

import (
	"context"
	"errors"
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/jackc/pgx/v5"
)

func (store *Store) NextAnalyticsDashboardExportMaintenanceUnknown(
	ctx context.Context,
) (platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate, bool, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		return platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportSourceUnavailable
	}
	candidate := platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{}
	var requestDigest []byte
	err := store.database.QueryRow(ctx, maintenanceUnknownCandidateSQL).Scan(
		&candidate.Kind, &candidate.TenantID, &candidate.TenantSchema,
		&candidate.ExportID, &candidate.AuthorityGroupID,
		&candidate.AuthorityGroupGeneration, &candidate.CutoverGeneration,
		&candidate.JobID, &candidate.AttemptID, &candidate.AttemptNumber,
		&candidate.LeaseGeneration, &candidate.OriginalLeaseOwnerID,
		&candidate.EffectID, &requestDigest, &candidate.UnknownAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{}, false, nil
	}
	if err != nil {
		return platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{}, false,
			fmt.Errorf("read analytics dashboard export maintenance unknown candidate: %w", err)
	}
	if len(requestDigest) != len(candidate.EffectRequestSHA256) {
		return platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportEvidenceInvalid
	}
	copy(candidate.EffectRequestSHA256[:], requestDigest)
	if !candidate.Valid() {
		return platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportEvidenceInvalid
	}
	return candidate, true, nil
}
