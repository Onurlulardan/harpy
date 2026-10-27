package upload

import (
	"context"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

func (storage *store) PersistEffect(ctx context.Context,
	receipt uploadapp.EffectReceipt,
) error {
	if !storage.usable(ctx) || !storage.command || !validReceipt(storage, receipt) {
		return uploadapp.ErrInvalid
	}
	var recorded bool
	err := storage.queryRow(ctx, recordEffectSQL, receipt.Plan.UploadID.String(),
		storage.actorID.String(), receipt.Plan.Generation, receipt.Object.Key,
		receipt.Object.Version, receipt.Object.Generation, receipt.Object.PlaintextBytes,
		receipt.Object.CiphertextBytes, receipt.Object.PlaintextSHA256[:],
		receipt.Object.CiphertextSHA256[:], receipt.DetectedMIME,
		string(receipt.DetectedKind), receipt.Object.KeyID, receipt.Object.WrappedDEK,
		receipt.ScanRevision, receipt.ScannedAt).Scan(&recorded)
	if err != nil {
		if storage.notFound(err) {
			return uploadapp.ErrConflict
		}
		return uploadapp.ErrUnavailable
	}
	if !recorded {
		return uploadapp.ErrConflict
	}
	return nil
}

func validReceipt(storage *store, receipt uploadapp.EffectReceipt) bool {
	plan := receipt.Plan
	object := receipt.Object
	return plan.ActorID == storage.actorID && plan.TenantID == storage.tenantID &&
		plan.UploadID.UUIDv7() && plan.DocumentID.UUIDv7() &&
		plan.CandidateVersionID.UUIDv7() && plan.Generation > 0 &&
		object.Key == plan.ObjectKey && object.Version != "" &&
		object.Generation == plan.Generation && object.PlaintextBytes == plan.ExpectedBytes &&
		object.CiphertextBytes > object.PlaintextBytes && object.KeyID != "" &&
		len(object.WrappedDEK) >= 115 && receipt.DetectedMIME != "" &&
		receipt.DetectedKind == plan.DeclaredKind && receipt.ScanRevision != "" &&
		validTime(receipt.ScannedAt)
}
