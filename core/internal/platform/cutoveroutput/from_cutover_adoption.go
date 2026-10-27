package cutoveroutput

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

func FromAdoption(
	operationID string,
	runID string,
	batchSHA256 string,
	appliedCount int64,
	keyStateSHA256 string,
	stateSHA256 string,
) (Result, error) {
	return newResult(runtimev1.NewCutoverAdoptionCommandResult(
		operationID, runID, batchSHA256, appliedCount, keyStateSHA256, stateSHA256,
	))
}
