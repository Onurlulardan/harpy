package runtimev1

import "fmt"

func validateCutoverAdoptionResult(result CutoverCommandResult) error {
	if result.Status != "succeeded" || result.ExitCode != ExitCodeClean ||
		result.BatchSHA256 == nil || !validMigratorDigest(*result.BatchSHA256) ||
		result.AppliedCount == nil || *result.AppliedCount < 0 ||
		result.KeyStateSHA256 == nil || !validMigratorDigest(*result.KeyStateSHA256) ||
		result.StateSHA256 == nil || !validMigratorDigest(*result.StateSHA256) ||
		result.TerminalState != nil || result.TerminalEvidenceSHA256 != nil {
		return fmt.Errorf("cutover adoption result is invalid")
	}
	return nil
}

func validateCutoverFinalizationResult(result CutoverCommandResult) error {
	if result.TerminalState == nil || result.TerminalEvidenceSHA256 == nil ||
		!validMigratorDigest(*result.TerminalEvidenceSHA256) ||
		result.BatchSHA256 != nil || result.AppliedCount != nil ||
		result.KeyStateSHA256 != nil || result.StateSHA256 != nil {
		return fmt.Errorf("cutover finalization result is invalid")
	}
	status, code, err := finalizationProcessOutcome(*result.TerminalState)
	if err != nil || result.Status != status || result.ExitCode != code {
		return fmt.Errorf("cutover finalization outcome is invalid")
	}
	return nil
}

func finalizationProcessOutcome(terminalState string) (string, ExitCode, error) {
	switch terminalState {
	case "completed":
		return "succeeded", ExitCodeClean, nil
	case "failed":
		return "failed", ExitCodeDeployMigration, nil
	case "cancelled":
		return "cancelled", ExitCodeDeployMigration, nil
	default:
		return "", ExitCodeDeployMigration,
			fmt.Errorf("cutover finalization terminal state is invalid")
	}
}
