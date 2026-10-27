package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *WorkerStore) FinalizeExpired(
	ctx context.Context, candidate exportapp.ExpiredCandidate, observedAt time.Time,
) (exportapp.ExpiredFinalization, error) {
	if !validExpiredCandidate(candidate) || observedAt.IsZero() {
		return exportapp.ExpiredFinalization{}, exportapp.ErrInvalid
	}
	var state string
	var finalizedAt time.Time
	err := storage.inTenant(ctx, candidate.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			schema, err := bound.currentSchema(boundContext)
			if err != nil || schema != candidate.TenantSchema {
				return exportapp.ErrUnavailable
			}
			var effectID, requestDigest any
			if candidate.EffectStarted {
				effectID = candidate.EffectID
				requestDigest = candidate.EffectRequestSHA256[:]
			}
			return bound.queryRow(boundContext, workerExpiredFinalizeSQL,
				schema, candidate.TenantID, candidate.ExportID, candidate.JobID,
				candidate.AttemptID, candidate.AttemptGeneration,
				candidate.LeaseOwnerID, effectID, requestDigest, observedAt,
			).Scan(&state, &finalizedAt)
		})
	if err != nil {
		return exportapp.ExpiredFinalization{}, mapDispatchError(err)
	}
	result := exportapp.ExpiredFinalization{FinalizedAt: finalizedAt.UTC()}
	switch state {
	case "outcome_unknown":
		result.State = exportapp.RenderExecutionOutcomeUnknown
	case "queued":
		result.State = exportapp.RenderExecutionRequeued
	case "dead":
		result.State = exportapp.RenderExecutionDead
	case "cancelled":
		result.State = exportapp.RenderExecutionCancelled
	default:
		return exportapp.ExpiredFinalization{}, exportapp.ErrUnavailable
	}
	if result.FinalizedAt.IsZero() ||
		(result.State == exportapp.RenderExecutionOutcomeUnknown) !=
			candidate.EffectStarted {
		return exportapp.ExpiredFinalization{}, exportapp.ErrUnavailable
	}
	return result, nil
}
