package cutoveroutput

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

func FromFinalization(
	operationID string,
	runID string,
	terminalState string,
	terminalEvidenceSHA256 string,
) (Result, error) {
	return newResult(runtimev1.NewCutoverFinalizationCommandResult(
		operationID, runID, terminalState, terminalEvidenceSHA256,
	))
}
