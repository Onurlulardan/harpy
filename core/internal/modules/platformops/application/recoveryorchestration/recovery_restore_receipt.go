package recoveryorchestration

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

const restoreReceiptValidity = 30 * 24 * time.Hour

func (executor *databaseRestoreExecutor) restoreReceipt(
	claim ports.RestoreExecutionClaim,
	manifest recoveryset.ManifestPayload,
	controlLocator, scratchLocator string,
) ([]byte, string, error) {
	payload := recoveryset.RestoreReceiptPayload{
		Schema:      recoveryset.RestoreReceiptSchema,
		OperationID: claim.OperationID, RecoverySetID: claim.RecoverySetID,
		PlanSHA256: claim.PlanSHA256, ManifestSHA256: claim.ManifestSHA256,
		Identity:                      manifest.Identity,
		PreviousDatabaseSystemID:      claim.PreviousDatabaseSystemID,
		NewDatabaseSystemID:           claim.NewDatabaseSystemID,
		PreviousDatabaseLocatorSHA256: controlLocator,
		NewDatabaseLocatorSHA256:      scratchLocator,
		DrainAckSHA256:                claim.DrainAckSHA256,
		SwitchAckSHA256:               claim.SwitchAckSHA256,
		ObjectMerkleSHA256:            claim.ObjectMerkleSHA256,
		LocatorRemapSHA256:            claim.LocatorRemapSHA256,
		SafetyEvidenceSHA256:          claim.SafetyEvidenceSHA256,
		Seeded:                        false, OwnerChanged: false, TenantReplaced: false,
		SwitchOutcome: "switched_same_environment",
	}
	payloadBytes, err := recoveryset.BuildRestoreReceipt(payload)
	if err != nil {
		return nil, "", err
	}
	path := executor.artifacts.ReceiptPath
	if _, err = os.Lstat(path); err == nil {
		return executor.readExistingRestoreReceipt(payload)
	} else if !os.IsNotExist(err) {
		return nil, "", err
	}
	issuedAt := time.Now().UTC().Truncate(time.Microsecond)
	artifact, err := environmentrecoveryv1.Build(
		environmentrecoveryv1.Input{
			Purpose:       environmentrecoveryv1.PurposeRestoreReceipt,
			EnvironmentID: claim.EnvironmentID, OperationID: claim.OperationID,
			ArtifactID: claim.OperationID,
			Sequence:   executor.authority.planArtifact.Sequence(),
			IssuedAt:   issuedAt, ExpiresAt: issuedAt.Add(restoreReceiptValidity),
			Payload: payloadBytes,
		}, executor.authority.trust, executor.authority.receiptSigner,
	)
	if err != nil {
		return nil, "", err
	}
	document := artifact.CanonicalBytes()
	if err = writeRecoveryArtifactAtomic(path, document); err != nil {
		return nil, "", err
	}
	return document, artifact.SHA256(), nil
}

func (executor *databaseRestoreExecutor) readExistingRestoreReceipt(
	expected recoveryset.RestoreReceiptPayload,
) ([]byte, string, error) {
	document, err := readRecoveryArtifact(
		executor.artifacts.ReceiptPath,
		environmentrecoveryv1.MaximumArtifactBytes, false,
	)
	if err != nil {
		return nil, "", err
	}
	artifact, err := environmentrecoveryv1.Parse(
		document, executor.authority.trust,
		time.Now().UTC().Truncate(time.Microsecond),
	)
	if err != nil || artifact.Purpose() !=
		environmentrecoveryv1.PurposeRestoreReceipt ||
		artifact.EnvironmentID() != expected.Identity.EnvironmentID ||
		artifact.OperationID() != expected.OperationID ||
		artifact.ArtifactID() != expected.OperationID ||
		artifact.Sequence() != executor.authority.planArtifact.Sequence() {
		return nil, "", terminalRestore(err)
	}
	payload, err := recoveryset.ParseRestoreReceipt(artifact.PayloadBytes())
	if err != nil || payload != expected {
		return nil, "", terminalRestore(err)
	}
	digest := sha256.Sum256(document)
	return document, hex.EncodeToString(digest[:]), nil
}
