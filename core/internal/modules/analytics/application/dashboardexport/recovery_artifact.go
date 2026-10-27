package dashboardexport

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"time"
)

func (runner *RecoveryRunner) cleanupUnknownArtifact(
	ctx context.Context, claim UnknownCleanupClaim,
) unknownCleanupResult {
	if claim.Draft == nil {
		return unknownCleanupResult{err: ErrUnavailable}
	}
	reconciled, err := runner.artifacts.Reconcile(ctx, *claim.Draft)
	if err != nil {
		return unknownCleanupResult{err: err}
	}
	switch reconciled.State {
	case ReconciliationMissing:
		if reconciled.Receipt.ProviderRequestID == "" ||
			reconciled.EvidenceSHA256 == [32]byte{} || reconciled.VerifiedAt.IsZero() {
			return unknownCleanupResult{err: ErrUnavailable}
		}
		return unknownCleanupResult{kind: UnknownResolutionAbsent,
			evidence: UnknownResolutionEvidence{ObjectKey: claim.Draft.ObjectKey,
				ReconcileProviderRequestID: reconciled.Receipt.ProviderRequestID,
				ReconcileEvidenceSHA256:    reconciled.EvidenceSHA256,
				AbsenceSHA256:              reconciled.EvidenceSHA256,
				ReconciledAt:               reconciled.VerifiedAt.UTC()}}
	case ReconciliationSealed:
		if reconciled.Receipt.ObjectVersion == "" ||
			reconciled.Receipt.ProviderRequestID == "" ||
			reconciled.Receipt.UploadedAt.IsZero() ||
			reconciled.EvidenceSHA256 == [32]byte{} || reconciled.VerifiedAt.IsZero() ||
			reconciled.VerifiedAt.Before(reconciled.Receipt.UploadedAt) {
			return unknownCleanupResult{err: ErrUnavailable}
		}
		deleted, deleteErr := runner.artifacts.Delete(
			ctx, *claim.Draft, reconciled.Receipt.ObjectVersion)
		if deleteErr != nil {
			return unknownCleanupResult{err: deleteErr}
		}
		if deleted.ProviderRequestID == "" || deleted.AbsenceSHA256 == [32]byte{} ||
			deleted.DeleteResponseSHA256 == [32]byte{} || deleted.VerifiedAbsentAt.IsZero() ||
			deleted.VerifiedAbsentAt.Before(reconciled.VerifiedAt) {
			return unknownCleanupResult{err: ErrUnavailable}
		}
		return unknownCleanupResult{kind: UnknownResolutionDeleted,
			evidence: UnknownResolutionEvidence{ObjectKey: claim.Draft.ObjectKey,
				ObjectVersion:              reconciled.Receipt.ObjectVersion,
				ReconcileProviderRequestID: reconciled.Receipt.ProviderRequestID,
				DeleteProviderRequestID:    deleted.ProviderRequestID,
				ReconcileEvidenceSHA256:    reconciled.EvidenceSHA256,
				DeleteResponseSHA256:       deleted.DeleteResponseSHA256,
				AbsenceSHA256:              deleted.AbsenceSHA256,
				UploadedAt:                 reconciled.Receipt.UploadedAt.UTC(),
				ReconciledAt:               reconciled.VerifiedAt.UTC(),
				DeletedAt:                  deleted.VerifiedAbsentAt.UTC()}}
	case ReconciliationUnknown:
		return unknownCleanupResult{err: ErrOutcomeUnknown}
	default:
		return unknownCleanupResult{err: ErrUnavailable}
	}
}

func unknownResolutionDigest(kind UnknownResolutionKind,
	evidence UnknownResolutionEvidence, resolvedAt time.Time,
) [32]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("analytics-dashboard-export-unknown-resolution-v1"))
	for _, value := range [][]byte{[]byte(kind), []byte(evidence.ObjectKey),
		[]byte(evidence.ObjectVersion), []byte(evidence.ReconcileProviderRequestID),
		[]byte(evidence.DeleteProviderRequestID)} {
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(len(value)))
		_, _ = hash.Write(length)
		_, _ = hash.Write(value)
	}
	_, _ = hash.Write(evidence.ReconcileEvidenceSHA256[:])
	writeUnknownResolutionTime(hash, evidence.UploadedAt)
	writeUnknownResolutionTime(hash, evidence.ReconciledAt)
	_, _ = hash.Write(evidence.DeleteResponseSHA256[:])
	_, _ = hash.Write(evidence.AbsenceSHA256[:])
	writeUnknownResolutionTime(hash, evidence.DeletedAt)
	writeUnknownResolutionTime(hash, resolvedAt)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func writeUnknownResolutionTime(hash interface{ Write([]byte) (int, error) }, value time.Time) {
	encoded := make([]byte, 8)
	if !value.IsZero() {
		binary.BigEndian.PutUint64(encoded, uint64(value.UTC().UnixMicro()))
	}
	_, _ = hash.Write(encoded)
}
