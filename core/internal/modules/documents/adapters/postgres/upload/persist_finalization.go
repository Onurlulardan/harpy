package upload

import (
	"context"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) PersistFinalization(ctx context.Context, receipt uploadapp.EffectReceipt,
	operationID, auditID domain.ID,
) (uploadapp.ReadyResult, error) {
	if !storage.usable(ctx) || !storage.command || !validReceipt(storage, receipt) ||
		!operationID.UUIDv7() || !auditID.UUIDv7() {
		return uploadapp.ReadyResult{}, uploadapp.ErrInvalid
	}
	var status int16
	var response []byte
	err := storage.queryRow(ctx, finalizeSQL, receipt.Plan.UploadID.String(),
		storage.actorID.String(), receipt.Plan.Generation, receipt.Object.Key,
		receipt.Object.Version, receipt.Object.Generation, receipt.Object.PlaintextBytes,
		receipt.Object.CiphertextBytes, receipt.Object.PlaintextSHA256[:],
		receipt.Object.CiphertextSHA256[:], receipt.DetectedMIME,
		string(receipt.DetectedKind), receipt.Object.KeyID, receipt.Object.WrappedDEK,
		receipt.ScanRevision, receipt.ScannedAt, operationID.String(),
		auditID.String()).Scan(&status, &response)
	if err != nil {
		if storage.notFound(err) {
			return uploadapp.ReadyResult{}, uploadapp.ErrConflict
		}
		return uploadapp.ReadyResult{}, uploadapp.ErrUnavailable
	}
	snapshot, err := scanSnapshot(storage.queryRow(ctx, getSQL,
		receipt.Plan.UploadID.String(), storage.actorID.String()), storage.tenantID,
		storage.notFound)
	if err != nil || status != 201 || snapshot.State != domain.UploadReady {
		return uploadapp.ReadyResult{}, uploadapp.ErrUnavailable
	}
	return uploadapp.ReadyResult{Upload: snapshot, DocumentID: receipt.Plan.DocumentID,
		VersionID: receipt.Plan.CandidateVersionID, Status: int(status),
		Response: append([]byte(nil), response...)}, nil
}
