package recoveryorchestration

import (
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	restoreprogressv1 "github.com/Onurlulardan/harpy/contracts/restoreprogress/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

const restoreProgressValidity = 24 * time.Hour

func (executor *databaseRestoreExecutor) writeProgress(
	claim ports.RestoreExecutionClaim,
	state, failureCode string,
	controlLocator, scratchLocator string,
) error {
	occurredAt := time.Now().UTC().Truncate(time.Microsecond)
	payload, err := restoreprogressv1.Build(restoreprogressv1.Input{
		EnvironmentID: claim.EnvironmentID, OperationID: claim.OperationID,
		AttemptID: claim.AttemptID, PlanSHA256: claim.PlanSHA256,
		State: state, PreviousDatabaseSystemID: claim.PreviousDatabaseSystemID,
		NewDatabaseSystemID:           claim.NewDatabaseSystemID,
		PreviousDatabaseLocatorSHA256: controlLocator,
		NewDatabaseLocatorSHA256:      scratchLocator,
		DrainAckSHA256:                claim.DrainAckSHA256,
		SwitchAckSHA256:               claim.SwitchAckSHA256,
		FailureCode:                   failureCode, OccurredAt: occurredAt,
	})
	if err != nil {
		return err
	}
	artifact, err := environmentrecoveryv1.Build(
		environmentrecoveryv1.Input{
			Purpose:       environmentrecoveryv1.PurposeRestoreProgress,
			EnvironmentID: claim.EnvironmentID, OperationID: claim.OperationID,
			ArtifactID: claim.OperationID, Sequence: restoreProgressSequence(state),
			IssuedAt: occurredAt, ExpiresAt: occurredAt.Add(restoreProgressValidity),
			Payload: payload,
		},
		executor.authority.trust, executor.authority.receiptSigner,
	)
	if err != nil {
		return err
	}
	return writeRecoveryArtifactAtomic(
		executor.artifacts.ProgressPath, artifact.CanonicalBytes(),
	)
}

func restoreProgressSequence(state string) int64 {
	switch state {
	case restoreprogressv1.StateReadyToStop:
		return 1
	case restoreprogressv1.StateRestoring:
		return 2
	case restoreprogressv1.StateReadyToSwitch:
		return 3
	case restoreprogressv1.StateCompleted:
		return 4
	default:
		return 5
	}
}
