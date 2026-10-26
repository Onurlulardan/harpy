package application

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

const restorePlanLifetime = 15 * time.Minute

func (service *RecoveryService) RequestPlannedRestore(
	ctx context.Context,
	request api.PlannedRestoreRequest,
) (api.RestoreOperation, error) {
	if service == nil || service.planSigner == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(request.OperationID) || !validRecoveryUUID(request.RecoverySetID) ||
		!validRecoveryUUID(request.RequestedByTenantID) ||
		!validRecoveryUUID(request.RequestedByActorID) ||
		!validRecoveryUUID(request.RequestedBySessionID) {
		return api.RestoreOperation{}, api.ErrInvalid
	}
	existing, existingErr := service.restores.ResolveRestore(ctx, request.OperationID)
	if existingErr == nil {
		if existing.RecoverySetID != request.RecoverySetID {
			return api.RestoreOperation{}, api.ErrConflict
		}
		return viewFromRecord(existing), nil
	}
	if !errors.Is(existingErr, ports.ErrNotFound) {
		return api.RestoreOperation{}, mapRecoveryStoreError(existingErr)
	}
	set, err := service.sets.GetRecoverySet(ctx, request.RecoverySetID)
	if err != nil {
		return api.RestoreOperation{}, mapRecoveryStoreError(err)
	}
	if set.State != "protected" || set.Sequence < 1 {
		return api.RestoreOperation{}, api.ErrConflict
	}
	manifest, err := service.verifyManifest(ctx, set)
	if err != nil {
		return api.RestoreOperation{}, err
	}
	target, err := service.identities.CurrentIdentity(ctx, set.EnvironmentID)
	if err != nil {
		return api.RestoreOperation{}, mapRecoveryStoreError(err)
	}
	safety := recoveryset.SafetyEvidence{RecoverySetID: request.SafetyRecoverySetID,
		RecoverySetSHA256: request.SafetyRecoverySetSHA256,
		WaiverReason:      request.WaiverReason, ProbeSHA256: request.ProbeSHA256,
		DestructiveConfirmed: request.DestructiveConfirmed}
	gate := recoveryset.RestoreRequest{RecoverySetID: set.RecoverySetID,
		Manifest: manifest, Target: target.Identity,
		Disposition: recoveryset.TargetDisposition(request.Disposition), Safety: safety}
	payload, err := recoveryset.BuildRestorePlan(gate, set.RecoverySetID, set.ManifestSHA256)
	if err != nil {
		return api.RestoreOperation{}, api.ErrConflict
	}
	now := service.clock.Now()
	signed, err := service.planSigner.SignRestorePlan(ctx, ports.RestorePlanSigningRequest{
		EnvironmentID: set.EnvironmentID, OperationID: request.OperationID,
		Sequence: set.Sequence, Payload: payload, IssuedAt: now,
		ExpiresAt: now.Add(restorePlanLifetime),
	})
	if err != nil {
		return api.RestoreOperation{}, api.ErrUnavailable
	}
	return service.RequestRestore(ctx, api.RestoreOperationRequest{
		OperationID: request.OperationID, RecoverySetID: set.RecoverySetID,
		RequestedByTenantID:  request.RequestedByTenantID,
		RequestedByActorID:   request.RequestedByActorID,
		RequestedBySessionID: request.RequestedBySessionID,
		PlanBytes:            signed.Document, PlanSigningGeneration: signed.Generation,
	})
}
