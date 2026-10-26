package systemtransfer

import (
	"crypto/sha256"
	"encoding/binary"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

const ExportReconciliationMaximumAttempts = 8

type ExportReconciliationClaim struct {
	TenantID, ExportID uuid.UUID
	JobID, AttemptID   uuid.UUID
	LeaseOwnerID       uuid.UUID
	Resource           publicresource.Resource
	Attempt            int
	MaximumAttempts    int
	LeaseGeneration    int64
}

func (claim ExportReconciliationClaim) Validate() error {
	if claim.TenantID == uuid.Nil || !exportV7(claim.ExportID) ||
		!exportV7(claim.JobID) || !exportV7(claim.AttemptID) ||
		!exportV7(claim.LeaseOwnerID) || !claim.Resource.Valid() ||
		claim.Attempt < 1 || claim.Attempt > ExportReconciliationMaximumAttempts ||
		claim.MaximumAttempts != ExportReconciliationMaximumAttempts ||
		claim.LeaseGeneration != int64(claim.Attempt) {
		return ErrExportReconciliationInvalid
	}
	return nil
}

type ExportReconciliationObservation struct {
	ExportEffect    ExportArtifactEffectStart
	EffectID        uuid.UUID
	EffectReceiptID uuid.UUID
	RequestSHA256   [32]byte
	StartedAt       time.Time
}

func (value ExportReconciliationObservation) Validate(
	claim ExportReconciliationClaim,
) error {
	if claim.Validate() != nil || value.ExportEffect.Validate() != nil ||
		value.ExportEffect.Attempt.State != ExportOutcomeUnknown ||
		value.ExportEffect.Attempt.TenantID != claim.TenantID ||
		value.ExportEffect.Attempt.ExportID != claim.ExportID ||
		value.ExportEffect.Attempt.Resource != claim.Resource ||
		value.ExportEffect.Attempt.JobID == claim.JobID ||
		value.EffectID != claim.AttemptID || !exportV7(value.EffectReceiptID) ||
		value.EffectReceiptID == value.EffectID ||
		value.EffectReceiptID == value.ExportEffect.EffectReceiptID ||
		value.RequestSHA256 != exportReconciliationRequestSHA(value.ExportEffect) ||
		!exportArtifactTime(value.StartedAt) ||
		value.StartedAt.Before(value.ExportEffect.StartedAt) {
		return ErrExportReconciliationInvalid
	}
	return nil
}

func exportReconciliationRequestSHA(effect ExportArtifactEffectStart) [32]byte {
	hash := sha256.New()
	hash.Write([]byte("iam.system-export.reconciliation.v1\x00"))
	hash.Write(effect.EffectID[:])
	hash.Write(effect.RequestSHA256[:])
	hash.Write(effect.Draft.CiphertextSHA256[:])
	hash.Write(effect.Draft.HeaderSHA256[:])
	writeExportReconcileString(hash, effect.ObjectKey)
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(effect.Draft.CiphertextBytes))
	hash.Write(encoded[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func writeExportReconcileString(hash interface{ Write([]byte) (int, error) }, value string) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(len(value)))
	_, _ = hash.Write(encoded[:])
	_, _ = hash.Write([]byte(value))
}
