package recoveryorchestration

import (
	"context"

	restoreorchestrationv1 "github.com/Onurlulardan/harpy/contracts/restoreorchestration/v1"
	restoreprogressv1 "github.com/Onurlulardan/harpy/contracts/restoreprogress/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseRestoreExecutor) enterRestoreSwitch(
	ctx context.Context,
	claim *ports.RestoreExecutionClaim,
	identity ports.RestoreExecutionIdentity,
	evidence databaseRestoreEvidence,
	controlLocator, scratchLocator string,
) error {
	claim.NewDatabaseSystemID = evidence.newDatabaseID
	if err := executor.writeProgress(
		*claim, restoreprogressv1.StateReadyToSwitch, "",
		controlLocator, scratchLocator,
	); err != nil {
		return resumableRestore(err)
	}
	ack, err := executor.waitForAck(
		ctx, executor.artifacts.SwitchAckPath,
		restoreorchestrationv1.PurposeSwitch, *claim,
		controlLocator, scratchLocator,
	)
	if err != nil {
		return err
	}
	record := ports.RestoreSwitchEvidence{
		RestoreExecutionIdentity:     identity,
		ObjectMerkleSHA256:           evidence.objectMerkle,
		LocatorRemapSHA256:           evidence.locatorRemap,
		SafetyEvidenceSHA256:         evidence.safety,
		SwitchAckSHA256:              ack.SHA256(),
		PreviousDatabaseSystemID:     claim.PreviousDatabaseSystemID,
		NewDatabaseSystemID:          evidence.newDatabaseID,
		ControlDatabaseLocatorSHA256: controlLocator,
		ScratchDatabaseLocatorSHA256: scratchLocator,
	}
	if err = executor.execution.MarkSwitched(ctx, record); err != nil {
		return resumableRestore(err)
	}
	claim.RestoreState = "switched"
	claim.SwitchAckSHA256 = ack.SHA256()
	claim.ObjectMerkleSHA256 = evidence.objectMerkle
	claim.LocatorRemapSHA256 = evidence.locatorRemap
	claim.SafetyEvidenceSHA256 = evidence.safety
	return nil
}
