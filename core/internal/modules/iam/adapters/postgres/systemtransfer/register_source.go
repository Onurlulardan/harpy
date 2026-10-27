package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func (storage *store) RegisterSource(ctx context.Context,
	write app.SourceWrite) (app.SourceRecord, error) {
	if !storage.usable(ctx) || !storage.command || write.Validate() != nil ||
		write.Evidence.TenantID != storage.tenantID ||
		write.Evidence.ActorID != storage.actorID {
		return app.SourceRecord{}, app.ErrControlUnavailable
	}
	artifact := write.Artifact
	var result app.SourceRecord
	var replay bool
	err := storage.queryRow(ctx, registerSourceSQL, write.SourceID, write.ReceiptID,
		storage.actorID, string(write.Evidence.Resource), write.Evidence.TokenVersion,
		write.Evidence.AuthorizationVersion, write.IdempotencyKeySHA256[:], write.Format,
		artifact.ObjectKey, artifact.ObjectVersion, artifact.HeaderSHA256[:],
		artifact.CiphertextSHA256[:], artifact.CiphertextBytes,
		artifact.PlaintextSHA256[:], artifact.PlaintextBytes, artifact.KeyID,
		artifact.WrappedDEK, artifact.WrapNonce[:], artifact.ProviderRequestID,
		artifact.UploadedAt).Scan(&result.SourceID, &result.Resource, &result.Format,
		&result.PlaintextBytes, &result.PlaintextSHA256, &result.UploadedAt, &replay)
	if err != nil {
		return app.SourceRecord{}, mapError(err, storage.notFound)
	}
	result.Replay = replay
	if result.SourceID == uuid.Nil {
		return app.SourceRecord{}, app.ErrControlUnavailable
	}
	return result, nil
}
