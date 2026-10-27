package recoveryorchestration

import (
	"context"

	restoreprogressv1 "github.com/Onurlulardan/harpy/contracts/restoreprogress/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseRestoreExecutor) completeRestore(
	ctx context.Context,
	claim ports.RestoreExecutionClaim,
	identity ports.RestoreExecutionIdentity,
	manifest recoveryset.ManifestPayload,
	controlLocator, scratchLocator string,
) error {
	receipt, receiptSHA256, err := executor.restoreReceipt(
		claim, manifest, controlLocator, scratchLocator,
	)
	if err != nil {
		return resumableRestore(err)
	}
	digest, digestErr := decodeTenantSetDigest(manifest.Identity.TenantSetSHA256)
	if digestErr != nil {
		return terminalRestore(digestErr)
	}
	scratch, err := executor.openScratch(
		ctx, manifest.Identity.EnvironmentID,
		manifest.Identity.PrimaryBindingID, digest, manifest.DatabaseSystemID,
	)
	if err != nil {
		return dependencyRestore(err)
	}
	defer scratch.Close()
	record := ports.RestoreReceiptImport{
		OperationID: claim.OperationID, RecoverySetID: claim.RecoverySetID,
		EnvironmentID: claim.EnvironmentID, PlanSHA256: claim.PlanSHA256,
		ManifestSHA256: claim.ManifestSHA256, ReceiptSHA256: receiptSHA256,
		ReceiptBytes:                 receipt,
		PreviousDatabaseSystemID:     claim.PreviousDatabaseSystemID,
		NewDatabaseSystemID:          claim.NewDatabaseSystemID,
		ObjectMerkleSHA256:           claim.ObjectMerkleSHA256,
		LocatorRemapSHA256:           claim.LocatorRemapSHA256,
		SafetyEvidenceSHA256:         claim.SafetyEvidenceSHA256,
		DrainAckSHA256:               claim.DrainAckSHA256,
		SwitchAckSHA256:              claim.SwitchAckSHA256,
		ControlDatabaseLocatorSHA256: controlLocator,
		ScratchDatabaseLocatorSHA256: scratchLocator,
		ComponentInventory:           claim.ComponentInventory,
	}
	if err = scratch.ImportReceipt(ctx, record); err != nil {
		return resumableRestore(err)
	}
	if err = executor.writeProgress(
		claim, restoreprogressv1.StateCompleted, "",
		controlLocator, scratchLocator,
	); err != nil {
		return resumableRestore(err)
	}
	if err = executor.execution.Complete(ctx, ports.RestoreCompletion{
		RestoreExecutionIdentity: identity,
		ReceiptSHA256:            receiptSHA256,
		ReceiptBytes:             receipt,
	}); err != nil {
		return resumableRestore(err)
	}
	return nil
}
