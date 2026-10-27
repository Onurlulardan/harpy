package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *WorkerStore) CheckpointRender(ctx context.Context,
	attempt exportapp.RenderDispatchAttempt, observedAt time.Time,
) (exportapp.RenderLease, error) {
	if !validWorkerCandidate(attempt.Candidate) || attempt.AttemptID.Version() != 7 ||
		attempt.Claim.LeaseOwner.Version() != 7 || observedAt.IsZero() ||
		attempt.Claim.JobID != attempt.Candidate.JobID ||
		attempt.Claim.Generation != attempt.Candidate.AttemptGeneration {
		return exportapp.RenderLease{}, exportapp.ErrInvalid
	}
	var result exportapp.RenderLease
	err := storage.inTenant(ctx, attempt.Candidate.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			if bound.currentSchema == nil {
				return exportapp.ErrUnavailable
			}
			schema, err := bound.currentSchema(boundContext)
			if err != nil || schema != attempt.Candidate.TenantSchema {
				return exportapp.ErrDispatchCandidateChanged
			}
			return bound.queryRow(boundContext, workerCheckpointSQL, schema,
				attempt.Candidate.TenantID, attempt.Candidate.ExportID,
				attempt.Candidate.JobID, attempt.AttemptID,
				attempt.Candidate.AttemptGeneration, attempt.Claim.LeaseOwner,
				observedAt).Scan(&result.LeaseUntil, &result.CancellationRequested)
		})
	if err != nil {
		return exportapp.RenderLease{}, mapDispatchError(err)
	}
	if result.LeaseUntil.IsZero() || !result.LeaseUntil.After(observedAt) {
		return exportapp.RenderLease{}, exportapp.ErrUnavailable
	}
	return result, nil
}
