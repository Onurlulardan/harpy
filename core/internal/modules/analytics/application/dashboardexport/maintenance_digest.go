package dashboardexport

import (
	"crypto/sha256"
	"encoding/binary"
	"time"
)

func maintenanceResolutionDigest(kind UnknownResolutionKind, state State,
	evidence MaintenanceResolutionEvidence, resolvedAt time.Time,
) [32]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("analytics-dashboard-export-maintenance-resolution-v1"))
	for _, value := range []string{string(kind), string(state), evidence.ObjectKey,
		evidence.ObjectVersion, evidence.ReconcileProviderRequestID,
		evidence.DeleteProviderRequestID} {
		encoded := []byte(value)
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(len(encoded)))
		_, _ = hash.Write(length)
		_, _ = hash.Write(encoded)
	}
	for _, digest := range [][32]byte{evidence.ReconcileEvidenceSHA256,
		evidence.DeleteResponseSHA256, evidence.AbsenceSHA256,
		evidence.LocalStateSHA256} {
		_, _ = hash.Write(digest[:])
	}
	for _, observedAt := range []time.Time{evidence.UploadedAt, evidence.ReconciledAt,
		evidence.DeletedAt, evidence.LocalStateAt, resolvedAt} {
		writeUnknownResolutionTime(hash, observedAt)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func maintenanceEvidenceTime(evidence MaintenanceResolutionEvidence) time.Time {
	result := evidence.LocalStateAt
	for _, candidate := range []time.Time{
		evidence.UploadedAt, evidence.ReconciledAt, evidence.DeletedAt,
	} {
		if candidate.After(result) {
			result = candidate
		}
	}
	return result
}
