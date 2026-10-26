package recoveryorchestration

import (
	"context"
	"fmt"
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recoverykitv1 "github.com/Onurlulardan/harpy/contracts/recoverykit/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	rehydraterequestv1 "github.com/Onurlulardan/harpy/contracts/rehydraterequest/v1"
	restoreorchestrationv1 "github.com/Onurlulardan/harpy/contracts/restoreorchestration/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

func loadRehydrateRestoreAuthority(
	ctx context.Context,
	artifacts RestoreArtifacts,
	storage RestoreStorage,
	externalTrust recoverysigningv1.Bundle,
	orchestrator restoreorchestrationv1.Trust,
	now time.Time,
) (databaseRestoreAuthority, error) {
	kitDocument, err := readRecoveryArtifact(
		artifacts.RecoveryKit, recoverykitv1.MaximumKitBytes, true,
	)
	if err != nil {
		return databaseRestoreAuthority{}, err
	}
	defer clear(kitDocument)
	kit, err := recoverykitv1.Parse(kitDocument)
	if err != nil {
		return databaseRestoreAuthority{}, err
	}
	trust, err := kit.TrustBundle()
	if err != nil || trust.SHA256() != externalTrust.SHA256() ||
		orchestrator.EnvironmentID() != kit.EnvironmentID() ||
		orchestrator.DeploymentBindingSHA256() !=
			kit.DeploymentBindingSHA256() {
		kit.Clear()
		return databaseRestoreAuthority{},
			fmt.Errorf("rehydrate authority binding differs")
	}
	requestDocument, err := readRecoveryArtifact(
		artifacts.RehydrateRequest,
		rehydraterequestv1.MaximumDocumentBytes, false,
	)
	if err != nil {
		kit.Clear()
		return databaseRestoreAuthority{}, err
	}
	request, err := rehydraterequestv1.Parse(
		requestDocument, kitDocument, now,
	)
	if err != nil || request.ExpectedEnvironmentID != kit.EnvironmentID() {
		kit.Clear()
		return databaseRestoreAuthority{}, fmt.Errorf(
			"rehydrate request identity differs",
		)
	}
	manifest, err := discoverManifest(
		ctx, storage, kit.EnvironmentID(), request.RecoverySetID, trust,
	)
	if err != nil {
		kit.Clear()
		return databaseRestoreAuthority{}, err
	}
	planPayload, err := buildRehydratePlan(request, manifest)
	if err != nil {
		kit.Clear()
		return databaseRestoreAuthority{}, err
	}
	controlKey, found := kit.SigningPrivateKey(
		recoverysigningv1.PurposeRecoveryControl,
	)
	if !found {
		kit.Clear()
		return databaseRestoreAuthority{}, fmt.Errorf(
			"rehydrate control key is unavailable",
		)
	}
	defer clear(controlKey)
	planArtifact, err := environmentrecoveryv1.Build(
		environmentrecoveryv1.Input{
			Purpose:       environmentrecoveryv1.PurposeRestorePlan,
			EnvironmentID: kit.EnvironmentID(), OperationID: kit.OperationID(),
			ArtifactID: kit.OperationID(), Sequence: manifest.artifact.Sequence(),
			IssuedAt: now, ExpiresAt: now.Add(time.Hour), Payload: planPayload,
		}, trust, controlKey,
	)
	if err != nil {
		kit.Clear()
		return databaseRestoreAuthority{}, err
	}
	receiptKey, found := kit.SigningPrivateKey(
		recoverysigningv1.PurposeRestoreReceipt,
	)
	if !found {
		kit.Clear()
		return databaseRestoreAuthority{}, fmt.Errorf(
			"rehydrate receipt key is unavailable",
		)
	}
	plan, _ := recoveryset.ParseRestorePlan(planPayload)
	return databaseRestoreAuthority{
		trust: trust, orchestrator: orchestrator,
		planArtifact: planArtifact, plan: plan, manifest: manifest,
		receiptSigner: receiptKey, kit: &kit,
	}, nil
}

func buildRehydratePlan(
	request rehydraterequestv1.Request,
	manifest discoveredRecoveryManifest,
) ([]byte, error) {
	safety := recoveryset.SafetyEvidence{
		WaiverReason: request.WaiverReason, ProbeSHA256: request.ProbeSHA256,
		DestructiveConfirmed: request.DestructiveConfirmed,
	}
	restore := recoveryset.RestoreRequest{
		RecoverySetID: request.RecoverySetID, Manifest: manifest.payload,
		Target:      manifest.payload.Identity,
		Disposition: recoveryset.TargetDisposition(request.TargetDisposition),
		Safety:      safety,
	}
	return recoveryset.BuildRestorePlan(
		restore, request.RecoverySetID, manifest.artifact.SHA256(),
	)
}
