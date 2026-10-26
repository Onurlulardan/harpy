package upload

import (
	"context"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) PersistFailure(ctx context.Context, failure uploadapp.EffectFailure,
	deletionID domain.ID,
) error {
	if !storage.usable(ctx) || !storage.command || !validFailure(storage, failure, deletionID) {
		return uploadapp.ErrInvalid
	}
	arguments := failureArguments(failure, deletionID)
	arguments = append([]any{failure.Plan.UploadID.String(), storage.actorID.String(),
		failure.Plan.Generation, string(failure.Kind), failure.StableCode,
		failure.OccurredAt}, arguments...)
	var recorded bool
	err := storage.queryRow(ctx, recordFailureSQL, arguments...).Scan(&recorded)
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

func validFailure(storage *store, failure uploadapp.EffectFailure,
	deletionID domain.ID,
) bool {
	plan := failure.Plan
	if plan.ActorID != storage.actorID || plan.TenantID != storage.tenantID ||
		!plan.UploadID.UUIDv7() || plan.Generation == 0 || failure.StableCode == "" ||
		!validTime(failure.OccurredAt) {
		return false
	}
	if failure.Object == nil {
		return !deletionID.Valid()
	}
	object := failure.Object
	return deletionID.UUIDv7() && object.Key == plan.ObjectKey && object.Version != "" &&
		object.Generation == plan.Generation && object.CiphertextBytes > object.PlaintextBytes &&
		object.PlaintextSHA256 != [32]byte{} && object.CiphertextSHA256 != [32]byte{}
}

func failureArguments(failure uploadapp.EffectFailure, deletionID domain.ID) []any {
	if failure.Object == nil {
		return []any{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	}
	object := failure.Object
	return []any{deletionID.String(), object.Key, object.Version, object.Generation,
		object.PlaintextBytes, object.CiphertextBytes, object.PlaintextSHA256[:],
		object.CiphertextSHA256[:], object.KeyID, object.WrappedDEK}
}
