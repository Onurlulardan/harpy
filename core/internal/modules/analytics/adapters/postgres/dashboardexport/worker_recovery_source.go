package dashboardexport

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *WorkerStore) NextUnknown(
	ctx context.Context,
) (exportapp.UnknownCandidate, bool, error) {
	if storage == nil || storage.source == nil || ctx == nil || ctx.Err() != nil {
		return exportapp.UnknownCandidate{}, false, exportapp.ErrUnavailable
	}
	stored, found, err := storage.source.NextAnalyticsDashboardExportUnknown(ctx)
	if err != nil {
		return exportapp.UnknownCandidate{}, false, exportapp.ErrUnavailable
	}
	if !found {
		return exportapp.UnknownCandidate{}, false, nil
	}
	candidate := exportapp.UnknownCandidate{TenantID: stored.TenantID,
		TenantSchema: stored.TenantSchema, ExportID: stored.ExportID,
		JobID: stored.JobID, AttemptID: stored.AttemptID,
		AttemptGeneration: stored.AttemptGeneration,
		LeaseOwnerID:      stored.LeaseOwnerID, EffectID: stored.EffectID,
		EffectRequestSHA256: stored.EffectRequestSHA256,
		UnknownAt:           stored.UnknownAt.UTC()}
	if !validUnknownCandidate(candidate) {
		return exportapp.UnknownCandidate{}, false, exportapp.ErrUnavailable
	}
	return candidate, true, nil
}

func (storage *WorkerStore) NextExpired(
	ctx context.Context,
) (exportapp.ExpiredCandidate, bool, error) {
	if storage == nil || storage.source == nil || ctx == nil || ctx.Err() != nil {
		return exportapp.ExpiredCandidate{}, false, exportapp.ErrUnavailable
	}
	stored, found, err := storage.source.NextAnalyticsDashboardExportExpired(ctx)
	if err != nil {
		return exportapp.ExpiredCandidate{}, false, exportapp.ErrUnavailable
	}
	if !found {
		return exportapp.ExpiredCandidate{}, false, nil
	}
	candidate := exportapp.ExpiredCandidate{TenantID: stored.TenantID,
		TenantSchema: stored.TenantSchema, ExportID: stored.ExportID,
		JobID: stored.JobID, AttemptID: stored.AttemptID,
		AttemptGeneration: stored.AttemptGeneration,
		LeaseOwnerID:      stored.LeaseOwnerID, EffectStarted: stored.EffectStarted,
		EffectID: stored.EffectID, EffectRequestSHA256: stored.EffectRequestSHA256,
		CancellationPending: stored.CancellationPending,
		LeaseExpiredAt:      stored.LeaseExpiredAt.UTC()}
	if !validExpiredCandidate(candidate) {
		return exportapp.ExpiredCandidate{}, false, exportapp.ErrUnavailable
	}
	return candidate, true, nil
}

func (storage *WorkerStore) NextMaintenanceUnknown(
	ctx context.Context,
) (exportapp.MaintenanceUnknownCandidate, bool, error) {
	if storage == nil || storage.source == nil || ctx == nil || ctx.Err() != nil {
		return exportapp.MaintenanceUnknownCandidate{}, false, exportapp.ErrUnavailable
	}
	stored, found, err := storage.source.NextAnalyticsDashboardExportMaintenanceUnknown(ctx)
	if err != nil {
		return exportapp.MaintenanceUnknownCandidate{}, false, exportapp.ErrUnavailable
	}
	if !found {
		return exportapp.MaintenanceUnknownCandidate{}, false, nil
	}
	candidate := exportapp.MaintenanceUnknownCandidate{
		Kind: stored.Kind, TenantID: stored.TenantID,
		TenantSchema: stored.TenantSchema, ExportID: stored.ExportID,
		AuthorityGroupID:         stored.AuthorityGroupID,
		AuthorityGroupGeneration: stored.AuthorityGroupGeneration,
		CutoverGeneration:        stored.CutoverGeneration, JobID: stored.JobID,
		AttemptID: stored.AttemptID, AttemptNumber: stored.AttemptNumber,
		LeaseGeneration:      stored.LeaseGeneration,
		OriginalLeaseOwnerID: stored.OriginalLeaseOwnerID,
		EffectID:             stored.EffectID, EffectRequestSHA256: stored.EffectRequestSHA256,
		UnknownAt: stored.UnknownAt.UTC(),
	}
	if !validMaintenanceCandidate(candidate) {
		return exportapp.MaintenanceUnknownCandidate{}, false, exportapp.ErrUnavailable
	}
	return candidate, true, nil
}

func validUnknownCandidate(candidate exportapp.UnknownCandidate) bool {
	return candidate.TenantID != [16]byte{} &&
		workerTenantSchemaPattern.MatchString(candidate.TenantSchema) &&
		candidate.ExportID.Version() == 7 && candidate.JobID.Version() == 7 &&
		candidate.AttemptID.Version() == 7 && candidate.LeaseOwnerID.Version() == 7 &&
		candidate.EffectID.Version() == 7 && candidate.EffectRequestSHA256 != [32]byte{} &&
		candidate.AttemptGeneration >= 1 &&
		candidate.AttemptGeneration <= exportapp.MaximumAttempts &&
		!candidate.UnknownAt.IsZero()
}

func validExpiredCandidate(candidate exportapp.ExpiredCandidate) bool {
	base := candidate.TenantID != [16]byte{} &&
		workerTenantSchemaPattern.MatchString(candidate.TenantSchema) &&
		candidate.ExportID.Version() == 7 && candidate.JobID.Version() == 7 &&
		candidate.AttemptID.Version() == 7 && candidate.LeaseOwnerID.Version() == 7 &&
		candidate.AttemptGeneration >= 1 &&
		candidate.AttemptGeneration <= exportapp.MaximumAttempts &&
		!candidate.LeaseExpiredAt.IsZero()
	if !base || !candidate.EffectStarted {
		return base && candidate.EffectID == [16]byte{} &&
			candidate.EffectRequestSHA256 == [32]byte{}
	}
	return candidate.EffectID.Version() == 7 &&
		candidate.EffectRequestSHA256 != [32]byte{}
}
