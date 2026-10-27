package runtimev1

import "fmt"

const (
	CutoverCommandResultVersion = "cutover-command-result.v1"
	CutoverActionAdoption       = "cutover_adoption"
	CutoverActionFinalization   = "cutover_finalization"
)

type CutoverCommandResult struct {
	ContractVersion        string   `json:"contract_version"`
	Action                 string   `json:"action"`
	OperationID            string   `json:"operation_id"`
	RunID                  string   `json:"run_id"`
	Status                 string   `json:"status"`
	ExitCode               ExitCode `json:"exit_code"`
	BatchSHA256            *string  `json:"batch_sha256,omitempty"`
	AppliedCount           *int64   `json:"applied_count,omitempty"`
	KeyStateSHA256         *string  `json:"key_state_sha256,omitempty"`
	StateSHA256            *string  `json:"state_sha256,omitempty"`
	TerminalState          *string  `json:"terminal_state,omitempty"`
	TerminalEvidenceSHA256 *string  `json:"terminal_evidence_sha256,omitempty"`
}

func NewCutoverAdoptionCommandResult(
	operationID string,
	runID string,
	batchSHA256 string,
	appliedCount int64,
	keyStateSHA256 string,
	stateSHA256 string,
) (CutoverCommandResult, error) {
	result := CutoverCommandResult{
		ContractVersion: CutoverCommandResultVersion,
		Action:          CutoverActionAdoption, OperationID: operationID, RunID: runID,
		Status: "succeeded", ExitCode: ExitCodeClean,
		BatchSHA256: cutoverString(batchSHA256), AppliedCount: &appliedCount,
		KeyStateSHA256: cutoverString(keyStateSHA256),
		StateSHA256:    cutoverString(stateSHA256),
	}
	return result, result.Validate()
}

func NewCutoverFinalizationCommandResult(
	operationID string,
	runID string,
	terminalState string,
	terminalEvidenceSHA256 string,
) (CutoverCommandResult, error) {
	status, code, err := finalizationProcessOutcome(terminalState)
	if err != nil {
		return CutoverCommandResult{}, err
	}
	result := CutoverCommandResult{
		ContractVersion: CutoverCommandResultVersion,
		Action:          CutoverActionFinalization, OperationID: operationID, RunID: runID,
		Status: status, ExitCode: code, TerminalState: cutoverString(terminalState),
		TerminalEvidenceSHA256: cutoverString(terminalEvidenceSHA256),
	}
	return result, result.Validate()
}

func (result CutoverCommandResult) Validate() error {
	if result.ContractVersion != CutoverCommandResultVersion ||
		!validMigratorUUID(result.OperationID) || !validMigratorUUID(result.RunID) {
		return fmt.Errorf("cutover command result identity is invalid")
	}
	switch result.Action {
	case CutoverActionAdoption:
		return validateCutoverAdoptionResult(result)
	case CutoverActionFinalization:
		return validateCutoverFinalizationResult(result)
	default:
		return fmt.Errorf("cutover command result action is invalid")
	}
}

func cutoverString(value string) *string {
	return &value
}
