package upload

import (
	"context"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/google/uuid"
)

func (storage *store) PrepareFinalize(ctx context.Context,
	receipt uploadapp.EffectReceipt,
) (uploadapp.FinalizeDecision, error) {
	if !storage.usable(ctx) || !storage.command || !validReceipt(storage, receipt) {
		return uploadapp.FinalizeDecision{}, uploadapp.ErrInvalid
	}
	var ready, conflict, cancelled bool
	var current *uuid.UUID
	err := storage.queryRow(ctx, prepareFinalizeSQL, receipt.Plan.UploadID.String(),
		storage.actorID.String(), receipt.Plan.Generation,
		receipt.Object.Version).Scan(&ready, &conflict, &cancelled, &current)
	if err != nil {
		if storage.notFound(err) {
			return uploadapp.FinalizeDecision{}, uploadapp.ErrConflict
		}
		return uploadapp.FinalizeDecision{}, uploadapp.ErrUnavailable
	}
	decision := uploadapp.FinalizeDecision{Ready: ready, Conflict: conflict,
		Cancelled: cancelled}
	if current != nil {
		decision.CurrentVersionID = domain.ID(*current)
	}
	return decision, nil
}
