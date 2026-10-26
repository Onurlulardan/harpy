package systemtransfer

import (
	"crypto/sha256"
	"encoding/binary"
	"unicode/utf8"
)

func newExportReconciliationReceipt(
	observation ExportReconciliationObservation,
	result ExportArtifactReconciliation,
) (ExportReconciliationReceipt, error) {
	receipt := ExportReconciliationReceipt{
		EffectID: observation.EffectID, EffectReceiptID: observation.EffectReceiptID,
		RequestSHA256: observation.RequestSHA256, ResultSHA256: result.EvidenceSHA256,
		ProviderRequestID: result.ProviderRequestID,
		State:             result.State, ObservedAt: result.CheckedAt,
	}
	if receipt.Validate(observation, result) != nil {
		return ExportReconciliationReceipt{}, ErrExportReconciliationInvalid
	}
	return receipt, nil
}

func (receipt ExportReconciliationReceipt) Validate(
	observation ExportReconciliationObservation,
	result ExportArtifactReconciliation,
) error {
	if receipt.EffectID != observation.EffectID ||
		receipt.EffectReceiptID != observation.EffectReceiptID ||
		receipt.RequestSHA256 != observation.RequestSHA256 ||
		receipt.ResultSHA256 != result.EvidenceSHA256 ||
		receipt.ProviderRequestID != result.ProviderRequestID ||
		len(receipt.ProviderRequestID) < 1 || len(receipt.ProviderRequestID) > 512 ||
		!utf8.ValidString(receipt.ProviderRequestID) || receipt.State != result.State ||
		(receipt.State != ExportArtifactSealed && receipt.State != ExportArtifactMissing) ||
		receipt.ObservedAt != result.CheckedAt || !exportArtifactTime(receipt.ObservedAt) {
		return ErrExportReconciliationInvalid
	}
	return nil
}

func exportReconciliationFailureSHA(
	claim ExportReconciliationClaim,
	observation ExportReconciliationObservation,
	code string,
) [32]byte {
	hash := sha256.New()
	hash.Write([]byte("iam.system-export.reconciliation-failure.v1\x00"))
	hash.Write(claim.JobID[:])
	hash.Write(observation.EffectID[:])
	writeExportReconcileString(hash, code)
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(claim.Attempt))
	hash.Write(encoded[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}
