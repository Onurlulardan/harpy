package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *WorkerStore) CheckpointUnknown(
	ctx context.Context, claim exportapp.UnknownCleanupClaim, observedAt time.Time,
) (exportapp.UnknownCleanupLease, error) {
	if !validCleanupClaim(claim) || observedAt.IsZero() {
		return exportapp.UnknownCleanupLease{}, exportapp.ErrInvalid
	}
	result := exportapp.UnknownCleanupLease{}
	err := storage.inTenant(ctx, claim.Candidate.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			schema, err := cleanupSchema(boundContext, bound, claim)
			if err != nil {
				return err
			}
			arguments := append(cleanupArguments(schema, claim), observedAt)
			return bound.queryRow(boundContext, workerUnknownCleanupCheckpointSQL,
				arguments...).Scan(&result.LeaseUntil, &result.CancellationRequested)
		})
	if err != nil {
		return exportapp.UnknownCleanupLease{}, mapDispatchError(err)
	}
	result.LeaseUntil = result.LeaseUntil.UTC()
	if !result.LeaseUntil.After(observedAt) {
		return exportapp.UnknownCleanupLease{}, exportapp.ErrUnavailable
	}
	return result, nil
}

func (storage *WorkerStore) DeferUnknown(
	ctx context.Context, claim exportapp.UnknownCleanupClaim, observedAt time.Time,
) (time.Time, error) {
	if !validCleanupClaim(claim) || observedAt.IsZero() {
		return time.Time{}, exportapp.ErrInvalid
	}
	var retryAt time.Time
	err := storage.inTenant(ctx, claim.Candidate.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			schema, err := cleanupSchema(boundContext, bound, claim)
			if err != nil {
				return err
			}
			arguments := append(cleanupArguments(schema, claim), observedAt)
			return bound.queryRow(boundContext, workerUnknownCleanupDeferSQL,
				arguments...).Scan(&retryAt)
		})
	if err != nil {
		return time.Time{}, mapDispatchError(err)
	}
	return retryAt.UTC(), nil
}

func (storage *WorkerStore) ResolveUnknown(
	ctx context.Context, claim exportapp.UnknownCleanupClaim,
	resolution exportapp.UnknownResolution,
) (exportapp.State, error) {
	if !validCleanupClaim(claim) || resolution.ReceiptID.Version() != 7 ||
		resolution.Kind != exportapp.UnknownResolutionDeleted &&
			resolution.Kind != exportapp.UnknownResolutionAbsent ||
		resolution.SHA256 == [32]byte{} || resolution.ResolvedAt.IsZero() ||
		!validUnknownResolutionEvidence(resolution) {
		return "", exportapp.ErrInvalid
	}
	var state string
	err := storage.inTenant(ctx, claim.Candidate.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			schema, err := cleanupSchema(boundContext, bound, claim)
			if err != nil {
				return err
			}
			arguments := append(cleanupArguments(schema, claim),
				resolution.ReceiptID, string(resolution.Kind),
				resolution.SHA256[:], resolution.Evidence.ObjectKey,
				optionalText(resolution.Evidence.ObjectVersion),
				resolution.Evidence.ReconcileProviderRequestID,
				resolution.Evidence.ReconcileEvidenceSHA256[:],
				optionalTime(resolution.Evidence.UploadedAt),
				resolution.Evidence.ReconciledAt,
				optionalText(resolution.Evidence.DeleteProviderRequestID),
				optionalDigest(resolution.Evidence.DeleteResponseSHA256),
				resolution.Evidence.AbsenceSHA256[:],
				optionalTime(resolution.Evidence.DeletedAt), resolution.ResolvedAt)
			return bound.queryRow(boundContext, workerUnknownCleanupResolveSQL,
				arguments...).Scan(&state)
		})
	if err != nil {
		return "", mapDispatchError(err)
	}
	result := exportapp.State(state)
	if result != exportapp.StateFailed && result != exportapp.StateCancelled &&
		result != exportapp.StateExpired {
		return "", exportapp.ErrUnavailable
	}
	return result, nil
}

func validUnknownResolutionEvidence(resolution exportapp.UnknownResolution) bool {
	evidence := resolution.Evidence
	if evidence.ObjectKey == "" || evidence.ReconcileProviderRequestID == "" ||
		evidence.ReconcileEvidenceSHA256 == [32]byte{} ||
		evidence.AbsenceSHA256 == [32]byte{} || evidence.ReconciledAt.IsZero() ||
		resolution.ResolvedAt.Before(evidence.ReconciledAt) {
		return false
	}
	if resolution.Kind == exportapp.UnknownResolutionAbsent {
		return evidence.ObjectVersion == "" && evidence.DeleteProviderRequestID == "" &&
			evidence.DeleteResponseSHA256 == [32]byte{} && evidence.UploadedAt.IsZero() &&
			evidence.DeletedAt.IsZero() &&
			evidence.AbsenceSHA256 == evidence.ReconcileEvidenceSHA256
	}
	return evidence.ObjectVersion != "" && evidence.DeleteProviderRequestID != "" &&
		evidence.DeleteResponseSHA256 != [32]byte{} && !evidence.UploadedAt.IsZero() &&
		!evidence.DeletedAt.IsZero() &&
		!evidence.ReconciledAt.Before(evidence.UploadedAt) &&
		!evidence.DeletedAt.Before(evidence.ReconciledAt) &&
		!resolution.ResolvedAt.Before(evidence.DeletedAt)
}

func optionalTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func cleanupSchema(ctx context.Context, bound *workerBound,
	claim exportapp.UnknownCleanupClaim,
) (string, error) {
	schema, err := bound.currentSchema(ctx)
	if err != nil || schema != claim.Candidate.TenantSchema {
		return "", exportapp.ErrUnavailable
	}
	return schema, nil
}

func cleanupArguments(schema string, claim exportapp.UnknownCleanupClaim) []any {
	candidate := claim.Candidate
	return []any{schema, candidate.TenantID, candidate.ExportID, candidate.JobID,
		candidate.AttemptID, candidate.AttemptGeneration, candidate.LeaseOwnerID,
		candidate.EffectID, candidate.EffectRequestSHA256[:],
		claim.CleanupGeneration, claim.CleanupLeaseOwnerID}
}

func validCleanupClaim(claim exportapp.UnknownCleanupClaim) bool {
	return claim.Disposition == exportapp.UnknownClaimed &&
		validUnknownCandidate(claim.Candidate) && claim.CleanupGeneration >= 1 &&
		claim.CleanupLeaseOwnerID.Version() == 7 &&
		claim.CleanupLeaseUntil.After(claim.Candidate.UnknownAt) && claim.Draft != nil
}
