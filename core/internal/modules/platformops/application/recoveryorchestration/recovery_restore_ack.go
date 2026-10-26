package recoveryorchestration

import (
	"context"
	"fmt"
	"os"
	"time"

	restoreorchestrationv1 "github.com/Onurlulardan/harpy/contracts/restoreorchestration/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

const restoreAckPoll = 2 * time.Second

func (executor *databaseRestoreExecutor) waitForAck(
	ctx context.Context,
	path string,
	purpose string,
	claim ports.RestoreExecutionClaim,
	controlLocator, scratchLocator string,
) (restoreorchestrationv1.Acknowledgement, error) {
	expected := restoreorchestrationv1.Expected{
		Purpose: purpose, EnvironmentID: claim.EnvironmentID,
		OperationID: claim.OperationID, AttemptID: claim.AttemptID,
		PlanSHA256:                    claim.PlanSHA256,
		PreviousDatabaseSystemID:      claim.PreviousDatabaseSystemID,
		PreviousDatabaseLocatorSHA256: controlLocator,
		NewDatabaseLocatorSHA256:      scratchLocator,
	}
	if purpose == restoreorchestrationv1.PurposeSwitch {
		expected.NewDatabaseSystemID = claim.NewDatabaseSystemID
	}
	for {
		document, found, err := readRestoreAcknowledgement(path)
		if err != nil {
			return restoreorchestrationv1.Acknowledgement{}, terminalRestore(err)
		}
		if found {
			ack, parseErr := restoreorchestrationv1.ParseAcknowledgement(
				document, executor.authority.orchestrator, expected,
				time.Now().UTC().Truncate(time.Microsecond),
			)
			if parseErr != nil {
				return restoreorchestrationv1.Acknowledgement{},
					terminalRestore(parseErr)
			}
			return ack, nil
		}
		timer := time.NewTimer(restoreAckPoll)
		select {
		case <-ctx.Done():
			timer.Stop()
			return restoreorchestrationv1.Acknowledgement{},
				resumableRestore(ctx.Err())
		case <-timer.C:
		}
	}
}

func readRestoreAcknowledgement(
	path string,
) ([]byte, bool, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil || !info.Mode().IsRegular() ||
		info.Mode()&os.ModeSymlink != 0 {
		return nil, false, fmt.Errorf("restore acknowledgement is invalid")
	}
	document, err := readRecoveryArtifact(
		path, restoreorchestrationv1.MaximumDocumentBytes, false,
	)
	if err != nil {
		return nil, false, err
	}
	return document, true, nil
}
