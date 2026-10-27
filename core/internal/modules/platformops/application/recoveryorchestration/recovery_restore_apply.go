package recoveryorchestration

import (
	"context"

	restoreorchestrationv1 "github.com/Onurlulardan/harpy/contracts/restoreorchestration/v1"
	restoreprogressv1 "github.com/Onurlulardan/harpy/contracts/restoreprogress/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseRestoreExecutor) enterRestoreApply(
	ctx context.Context,
	claim *ports.RestoreExecutionClaim,
	identity ports.RestoreExecutionIdentity,
	controlLocator, scratchLocator string,
) error {
	if err := executor.writeProgress(
		*claim, restoreprogressv1.StateReadyToStop, "",
		controlLocator, scratchLocator,
	); err != nil {
		return resumableRestore(err)
	}
	ack, err := executor.waitForAck(
		ctx, executor.artifacts.DrainAckPath,
		restoreorchestrationv1.PurposeDrain, *claim,
		controlLocator, scratchLocator,
	)
	if err != nil {
		return err
	}
	if err = executor.execution.ExecuteApply(ctx, ports.RestoreApplyEvidence{
		RestoreExecutionIdentity:     identity,
		DrainAckSHA256:               ack.SHA256(),
		ControlDatabaseLocatorSHA256: controlLocator,
		ScratchDatabaseLocatorSHA256: scratchLocator,
	}); err != nil {
		return resumableRestore(err)
	}
	claim.RestoreState = "restoring"
	claim.DrainAckSHA256 = ack.SHA256()
	claim.ControlDatabaseLocatorSHA256 = controlLocator
	claim.ScratchDatabaseLocatorSHA256 = scratchLocator
	return executor.writeProgress(
		*claim, restoreprogressv1.StateRestoring, "",
		controlLocator, scratchLocator,
	)
}
