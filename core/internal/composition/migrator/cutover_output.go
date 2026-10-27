package migrator

import (
	"io"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutoveroutput"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func writeAdoptionResult(
	output io.Writer,
	operationID string,
	runID string,
	batchSHA256 string,
	appliedCount int64,
	keyStateSHA256 string,
	stateSHA256 string,
) error {
	result, err := cutoveroutput.FromAdoption(
		operationID, runID, batchSHA256, appliedCount, keyStateSHA256, stateSHA256,
	)
	if err != nil {
		return err
	}
	return result.Encode(output)
}

func writeFinalizationResult(
	output io.Writer,
	operationID string,
	runID string,
	terminalState string,
	terminalEvidenceSHA256 string,
) (processcode.Code, error) {
	result, err := cutoveroutput.FromFinalization(
		operationID, runID, terminalState, terminalEvidenceSHA256,
	)
	if err != nil {
		return processcode.ServeFailure, err
	}
	if err := result.Encode(output); err != nil {
		return processcode.ServeFailure, err
	}
	return result.ExitCode(), nil
}
