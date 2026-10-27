package recoveryorchestration

import (
	"crypto/ed25519"
	"fmt"
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recoverykitv1 "github.com/Onurlulardan/harpy/contracts/recoverykit/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	restoreorchestrationv1 "github.com/Onurlulardan/harpy/contracts/restoreorchestration/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

type databaseRestoreAuthority struct {
	trust         recoverysigningv1.Bundle
	orchestrator  restoreorchestrationv1.Trust
	planArtifact  environmentrecoveryv1.Artifact
	plan          recoveryset.RestorePlanPayload
	manifest      discoveredRecoveryManifest
	receiptSigner ed25519.PrivateKey
	kit           *recoverykitv1.Kit
}

func loadRestoreTrust(
	artifacts RestoreArtifacts,
) (recoverysigningv1.Bundle, restoreorchestrationv1.Trust, error) {
	document, err := readRecoveryArtifact(
		artifacts.RecoveryTrustBundle,
		recoverysigningv1.MaximumBundleBytes, false,
	)
	if err != nil {
		return recoverysigningv1.Bundle{}, restoreorchestrationv1.Trust{}, err
	}
	trust, err := recoverysigningv1.Parse(document)
	if err != nil {
		return recoverysigningv1.Bundle{}, restoreorchestrationv1.Trust{}, err
	}
	orchestratorDocument, err := readRecoveryArtifact(
		artifacts.OrchestratorTrustBundle,
		restoreorchestrationv1.MaximumDocumentBytes, false,
	)
	if err != nil {
		return recoverysigningv1.Bundle{}, restoreorchestrationv1.Trust{}, err
	}
	orchestrator, err := restoreorchestrationv1.ParseTrust(orchestratorDocument)
	if err != nil || orchestrator.EnvironmentID() != trust.EnvironmentID() {
		return recoverysigningv1.Bundle{}, restoreorchestrationv1.Trust{},
			fmt.Errorf("restore orchestrator trust differs")
	}
	return trust, orchestrator, nil
}

func loadExistingRestoreAuthority(
	artifacts RestoreArtifacts,
	trust recoverysigningv1.Bundle,
	orchestrator restoreorchestrationv1.Trust,
	now time.Time,
) (databaseRestoreAuthority, error) {
	document, err := readRecoveryArtifact(
		artifacts.RestorePlan, environmentrecoveryv1.MaximumArtifactBytes, false,
	)
	if err != nil {
		return databaseRestoreAuthority{}, err
	}
	planArtifact, plan, err := parseRestorePlan(document, trust, now)
	if err != nil {
		return databaseRestoreAuthority{}, err
	}
	signer, err := readRecoverySigningKey(artifacts.ReceiptSigningKey)
	if err != nil {
		return databaseRestoreAuthority{}, err
	}
	if orchestrator.EnvironmentID() != plan.Identity.EnvironmentID {
		clear(signer)
		return databaseRestoreAuthority{},
			fmt.Errorf("restore orchestrator identity differs")
	}
	return databaseRestoreAuthority{
		trust: trust, orchestrator: orchestrator,
		planArtifact: planArtifact, plan: plan, receiptSigner: signer,
	}, nil
}

func parseRestorePlan(
	document []byte,
	trust recoverysigningv1.Bundle,
	now time.Time,
) (environmentrecoveryv1.Artifact, recoveryset.RestorePlanPayload, error) {
	artifact, err := environmentrecoveryv1.Parse(document, trust, now)
	if err != nil || artifact.Purpose() != environmentrecoveryv1.PurposeRestorePlan ||
		artifact.EnvironmentID() != trust.EnvironmentID() ||
		artifact.ArtifactID() != artifact.OperationID() {
		return environmentrecoveryv1.Artifact{}, recoveryset.RestorePlanPayload{},
			fmt.Errorf("restore plan artifact is invalid")
	}
	plan, err := recoveryset.ParseRestorePlan(artifact.PayloadBytes())
	if err != nil || plan.Identity.EnvironmentID != artifact.EnvironmentID() {
		return environmentrecoveryv1.Artifact{}, recoveryset.RestorePlanPayload{},
			fmt.Errorf("restore plan payload is invalid")
	}
	return artifact, plan, nil
}

func (authority *databaseRestoreAuthority) clear() {
	if authority == nil {
		return
	}
	clear(authority.receiptSigner)
	if authority.kit != nil {
		authority.kit.Clear()
	}
}
