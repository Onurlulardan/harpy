package dashboardexport

import "context"

type maintenanceCleanupResult struct {
	kind     UnknownResolutionKind
	evidence MaintenanceResolutionEvidence
	err      error
}

func (runner *RecoveryRunner) cleanupMaintenanceArtifact(
	ctx context.Context, claim MaintenanceUnknownClaim,
) maintenanceCleanupResult {
	if claim.Draft == nil || claim.Draft.ObjectKey == "" || claim.ObjectVersion == "" {
		return maintenanceCleanupResult{err: ErrUnavailable}
	}
	reconciled, err := runner.artifacts.Reconcile(ctx, *claim.Draft)
	if err != nil {
		return maintenanceCleanupResult{err: err}
	}
	if reconciled.Receipt.ProviderRequestID == "" ||
		reconciled.EvidenceSHA256 == [32]byte{} || reconciled.VerifiedAt.IsZero() {
		return maintenanceCleanupResult{err: ErrUnavailable}
	}
	switch reconciled.State {
	case ReconciliationMissing:
		if !reconciled.Receipt.UploadedAt.IsZero() ||
			reconciled.Receipt.ObjectVersion != "" {
			return maintenanceCleanupResult{err: ErrUnavailable}
		}
	case ReconciliationSealed:
		if reconciled.Receipt.ObjectVersion != claim.ObjectVersion ||
			reconciled.Receipt.UploadedAt.IsZero() ||
			reconciled.VerifiedAt.Before(reconciled.Receipt.UploadedAt) {
			return maintenanceCleanupResult{err: ErrUnavailable}
		}
	case ReconciliationUnknown:
		return maintenanceCleanupResult{err: ErrOutcomeUnknown}
	default:
		return maintenanceCleanupResult{err: ErrUnavailable}
	}
	// Key-level lookup evidence never proves an exact historical object version is
	// absent. DeleteExact must verify the pinned version before terminalization.
	deleted, deleteErr := runner.artifacts.Delete(ctx, *claim.Draft, claim.ObjectVersion)
	if deleteErr != nil {
		return maintenanceCleanupResult{err: deleteErr}
	}
	if deleted.ProviderRequestID == "" || deleted.DeleteResponseSHA256 == [32]byte{} ||
		deleted.AbsenceSHA256 == [32]byte{} || deleted.VerifiedAbsentAt.IsZero() ||
		deleted.VerifiedAbsentAt.Before(reconciled.VerifiedAt) {
		return maintenanceCleanupResult{err: ErrUnavailable}
	}
	return maintenanceCleanupResult{kind: UnknownResolutionDeleted,
		evidence: MaintenanceResolutionEvidence{
			ObjectKey: claim.Draft.ObjectKey, ObjectVersion: claim.ObjectVersion,
			ReconcileProviderRequestID: reconciled.Receipt.ProviderRequestID,
			DeleteProviderRequestID:    deleted.ProviderRequestID,
			ReconcileEvidenceSHA256:    reconciled.EvidenceSHA256,
			DeleteResponseSHA256:       deleted.DeleteResponseSHA256,
			AbsenceSHA256:              deleted.AbsenceSHA256,
			UploadedAt:                 reconciled.Receipt.UploadedAt.UTC(),
			ReconciledAt:               reconciled.VerifiedAt.UTC(),
			DeletedAt:                  deleted.VerifiedAbsentAt.UTC()}}
}
