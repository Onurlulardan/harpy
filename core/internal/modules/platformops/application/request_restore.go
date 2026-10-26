package application

import (
	"context"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *RecoveryService) RequestRestore(
	ctx context.Context,
	request api.RestoreOperationRequest,
) (api.RestoreOperation, error) {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(request.OperationID) || !validRecoveryUUID(request.RecoverySetID) ||
		!validRecoveryUUID(request.RequestedByTenantID) ||
		!validRecoveryUUID(request.RequestedByActorID) ||
		!validRecoveryUUID(request.RequestedBySessionID) ||
		len(request.PlanBytes) == 0 ||
		request.PlanSigningGeneration < 1 {
		return api.RestoreOperation{}, api.ErrInvalid
	}
	set, err := service.sets.GetRecoverySet(ctx, request.RecoverySetID)
	if err != nil {
		return api.RestoreOperation{}, mapRecoveryStoreError(err)
	}
	if set.State != "protected" || set.ManifestSHA256 == "" ||
		len(set.ManifestBytes) == 0 || set.RecoverySigningGeneration < 1 {
		return api.RestoreOperation{}, api.ErrConflict
	}
	manifest, err := service.verifyManifest(ctx, set)
	if err != nil {
		return api.RestoreOperation{}, err
	}
	plan, planArtifact, err := service.verifyPlan(ctx, request, set.EnvironmentID)
	if err != nil {
		return api.RestoreOperation{}, err
	}
	if plan.RecoverySetID != set.RecoverySetID ||
		plan.ManifestSHA256 != set.ManifestSHA256 ||
		planArtifact.OperationID() != request.OperationID ||
		planArtifact.ArtifactID() != request.OperationID {
		return api.RestoreOperation{}, api.ErrConflict
	}
	targetRecord, err := service.identities.CurrentIdentity(ctx, set.EnvironmentID)
	if err != nil {
		return api.RestoreOperation{}, mapRecoveryStoreError(err)
	}
	target := targetRecord.Identity
	safety := recoveryset.SafetyEvidence{RecoverySetID: plan.SafetyRecoverySetID,
		RecoverySetSHA256: plan.SafetyRecoverySetSHA256,
		WaiverReason:      plan.WaiverReason, ProbeSHA256: plan.ProbeSHA256,
		DestructiveConfirmed: plan.DestructiveConfirmed}
	gate := recoveryset.RestoreRequest{RecoverySetID: set.RecoverySetID,
		Manifest: manifest, Target: target,
		Disposition: plan.TargetDisposition, Safety: safety}
	if recoveryset.ValidateRestore(gate) != nil || plan.Identity != target {
		return api.RestoreOperation{}, api.ErrConflict
	}
	record := ports.RestoreRequestRecord{OperationID: request.OperationID,
		RecoverySetID: set.RecoverySetID, EnvironmentID: set.EnvironmentID,
		RequestedByTenantID:  request.RequestedByTenantID,
		RequestedByActorID:   request.RequestedByActorID,
		RequestedBySessionID: request.RequestedBySessionID,
		Disposition:          plan.TargetDisposition, PlanSHA256: planArtifact.SHA256(),
		PlanBytes: planArtifact.CanonicalBytes(), Safety: safety, CreatedAt: service.clock.Now()}
	if err = service.restores.RequestRestore(ctx, record); err != nil {
		return api.RestoreOperation{}, mapRecoveryStoreError(err)
	}
	return viewFromRecord(ports.RestoreRecord{OperationID: record.OperationID,
		RecoverySetID: record.RecoverySetID, EnvironmentID: record.EnvironmentID,
		Disposition: record.Disposition, State: "plan_ready", PlanSHA256: record.PlanSHA256,
		CreatedAt: record.CreatedAt, UpdatedAt: record.CreatedAt}), nil
}

func (service *RecoveryService) verifyManifest(
	ctx context.Context,
	set ports.RecoverySetRecord,
) (recoveryset.ManifestPayload, error) {
	bundle, err := service.loadSigningBundle(ctx, set.EnvironmentID,
		set.RecoverySigningGeneration, set.RecoveryTrustBundleSHA256)
	if err != nil {
		return recoveryset.ManifestPayload{}, err
	}
	artifact, err := environmentrecoveryv1.Parse(set.ManifestBytes, bundle, service.clock.Now())
	if err != nil || artifact.Purpose() != environmentrecoveryv1.PurposeRecoveryManifest ||
		artifact.EnvironmentID() != set.EnvironmentID || artifact.OperationID() != set.OperationID ||
		artifact.ArtifactID() != set.RecoverySetID || artifact.SHA256() != set.ManifestSHA256 {
		return recoveryset.ManifestPayload{}, api.ErrConflict
	}
	manifest, err := recoveryset.ParseManifest(artifact.PayloadBytes())
	if err != nil || manifest.DatabaseSystemID != set.DatabaseSystemID ||
		manifest.RecoverySigningGeneration != set.RecoverySigningGeneration ||
		manifest.RecoveryTrustBundleSHA256 != set.RecoveryTrustBundleSHA256 {
		return recoveryset.ManifestPayload{}, api.ErrConflict
	}
	return manifest, nil
}

func (service *RecoveryService) verifyPlan(
	ctx context.Context,
	request api.RestoreOperationRequest,
	environmentID string,
) (recoveryset.RestorePlanPayload, environmentrecoveryv1.Artifact, error) {
	bundle, err := service.loadSigningBundle(ctx, environmentID,
		request.PlanSigningGeneration, "")
	if err != nil {
		return recoveryset.RestorePlanPayload{}, environmentrecoveryv1.Artifact{}, err
	}
	artifact, err := environmentrecoveryv1.Parse(request.PlanBytes, bundle, service.clock.Now())
	if err != nil || artifact.Purpose() != environmentrecoveryv1.PurposeRestorePlan ||
		artifact.EnvironmentID() != environmentID {
		return recoveryset.RestorePlanPayload{}, environmentrecoveryv1.Artifact{}, api.ErrConflict
	}
	plan, err := recoveryset.ParseRestorePlan(artifact.PayloadBytes())
	if err != nil {
		return recoveryset.RestorePlanPayload{}, environmentrecoveryv1.Artifact{}, api.ErrConflict
	}
	return plan, artifact, nil
}

func (service *RecoveryService) loadSigningBundle(
	ctx context.Context, environmentID string, generation int64, expectedSHA string,
) (recoverysigningv1.Bundle, error) {
	record, err := service.signing.LoadSigningBundle(ctx, environmentID, generation)
	if err != nil {
		return recoverysigningv1.Bundle{}, mapRecoveryStoreError(err)
	}
	bundle, err := recoverysigningv1.Parse(record.BundleBytes)
	if err != nil || bundle.EnvironmentID() != environmentID ||
		bundle.Generation() != generation || bundle.SHA256() != record.SHA256 ||
		(expectedSHA != "" && bundle.SHA256() != expectedSHA) {
		return recoverysigningv1.Bundle{}, api.ErrConflict
	}
	return bundle, nil
}
