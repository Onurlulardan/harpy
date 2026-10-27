package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateStepRowV2(
	planHash string,
	step schemav2.Step,
	row StepRecord,
) (string, error) {
	if row.OperationID != string(step.OperationID) ||
		row.TransactionClass != string(step.TransactionClass) ||
		row.ExecutorKind != string(step.ExecutorKind) {
		return "", fmt.Errorf("persisted execution identity differs")
	}
	if err := compareFragment(row.IdentifierMapping, step.IdentifierMapping); err != nil {
		return "", fmt.Errorf("identifier mapping differs: %w", err)
	}
	if err := compareFragment(row.Arguments, step.Arguments); err != nil {
		return "", fmt.Errorf("arguments differ: %w", err)
	}
	prepared, err := schemav2.PrepareStep(planHash, step)
	if err != nil {
		return "", err
	}
	if row.StepSHA256 != prepared.SHA256() {
		return "", fmt.Errorf("step hash differs")
	}
	return prepared.SHA256(), nil
}
