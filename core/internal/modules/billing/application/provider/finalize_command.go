package providerapp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (finalizer *CommandFinalizer) FinalizeCommand(
	ctx context.Context,
	request FinalizeCommandRequest,
) (Finalization, error) {
	if finalizer == nil || finalizer.store == nil || finalizer.nextID == nil ||
		finalizer.clock == nil || ctx == nil || ctx.Err() != nil ||
		request.TenantID == uuid.Nil || request.JobID == uuid.Nil ||
		request.JobID.Version() != uuid.Version(7) || request.AttemptID == uuid.Nil ||
		request.CommandID == uuid.Nil || request.CommandID.Version() != uuid.Version(7) ||
		request.LeaseGeneration < 1 || request.PayloadRevision != 1 ||
		request.PayloadReference !=
			"billing-provider-finalize/"+request.CommandID.String() ||
		!lowerSHA256Pattern.MatchString(request.PayloadSHA256) {
		return Finalization{}, ErrCommandRequest
	}
	nextFinalizerID, err := nextCommandID(finalizer.nextID)
	if err != nil {
		return Finalization{}, err
	}
	if nextFinalizerID == request.JobID || nextFinalizerID == request.CommandID ||
		nextFinalizerID == request.AttemptID {
		return Finalization{}, ErrConfiguration
	}
	reconciliationID, err := nextCommandID(finalizer.nextID)
	if err != nil || reconciliationID == nextFinalizerID ||
		reconciliationID == request.JobID || reconciliationID == request.CommandID ||
		reconciliationID == request.AttemptID {
		return Finalization{}, ErrConfiguration
	}
	observedAt, err := canonicalInstant(finalizer.clock())
	if err != nil {
		return Finalization{}, ErrConfiguration
	}
	result, err := finalizer.store.FinalizeCommand(ctx, FinalizeCommandMutation{
		FinalizeCommandRequest: request,
		NextFinalizerJobID:     nextFinalizerID,
		ReconciliationJobID:    reconciliationID,
		ObservedAt:             observedAt,
	})
	if err != nil {
		return Finalization{}, fmt.Errorf("finalize billing provider command: %w", err)
	}
	rotatesFinalizer := result.Decision == FinalizePending ||
		result.Decision == FinalizeReconcile
	if !result.valid() || result.CommandID != request.CommandID ||
		(rotatesFinalizer && result.FinalizerJobID != nextFinalizerID) ||
		(!rotatesFinalizer && result.FinalizerJobID != request.JobID) ||
		(result.Decision == FinalizeReconcile && result.ActiveJobID != reconciliationID) {
		return Finalization{}, ErrCommandState
	}
	return result, nil
}
