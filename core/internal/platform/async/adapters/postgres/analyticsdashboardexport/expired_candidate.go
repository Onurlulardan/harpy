package analyticsdashboardexport

import (
	"context"
	"errors"
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (store *Store) NextAnalyticsDashboardExportExpired(
	ctx context.Context,
) (platformasync.AnalyticsDashboardExportExpiredCandidate, bool, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		return platformasync.AnalyticsDashboardExportExpiredCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportSourceUnavailable
	}
	candidate := platformasync.AnalyticsDashboardExportExpiredCandidate{}
	var effectID pgtype.UUID
	var requestDigest []byte
	err := store.database.QueryRow(ctx, expiredCandidateSQL).Scan(
		&candidate.TenantID, &candidate.TenantSchema, &candidate.ExportID,
		&candidate.JobID, &candidate.AttemptID, &candidate.AttemptGeneration,
		&candidate.LeaseOwnerID, &candidate.EffectStarted, &effectID,
		&requestDigest, &candidate.CancellationPending, &candidate.LeaseExpiredAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return platformasync.AnalyticsDashboardExportExpiredCandidate{}, false, nil
	}
	if err != nil {
		return platformasync.AnalyticsDashboardExportExpiredCandidate{}, false,
			fmt.Errorf("read analytics dashboard export expired candidate: %w", err)
	}
	if effectID.Valid {
		candidate.EffectID = uuid.UUID(effectID.Bytes)
	}
	if len(requestDigest) != 0 && len(requestDigest) != len(candidate.EffectRequestSHA256) {
		return platformasync.AnalyticsDashboardExportExpiredCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportEvidenceInvalid
	}
	copy(candidate.EffectRequestSHA256[:], requestDigest)
	if !candidate.Valid() {
		return platformasync.AnalyticsDashboardExportExpiredCandidate{}, false,
			platformasync.ErrAnalyticsDashboardExportEvidenceInvalid
	}
	return candidate, true, nil
}
