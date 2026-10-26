package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validatePlanSteps(plan schemav1.Plan, claim ClaimRecord) ([]ValidatedStep, error) {
	if len(claim.Steps) != len(plan.Steps) {
		return nil, fmt.Errorf("persisted step count differs")
	}
	validated := make([]ValidatedStep, len(plan.Steps))
	for ordinal, step := range plan.Steps {
		if err := validateResourceOperation(plan.Resource.Kind, step.OperationID); err != nil {
			return nil, err
		}
		row := claim.Steps[ordinal]
		contract, supported := operationContracts[step.OperationID]
		if !supported || step.Ordinal != ordinal || row.Ordinal != ordinal ||
			step.ExecutorKind != contract.executor || step.TransactionClass != contract.class {
			return nil, fmt.Errorf("step %d execution contract differs", ordinal)
		}
		if err := validateIdentifierMapping(step.IdentifierMapping, plan.Resource.ID); err != nil {
			return nil, fmt.Errorf("step %d identifiers: %w", ordinal, err)
		}
		arguments, err := decodeArguments(step)
		if err != nil {
			return nil, fmt.Errorf("step %d: %w", ordinal, err)
		}
		if err := validateOperationArguments(step, arguments); err != nil {
			return nil, fmt.Errorf("step %d: %w", ordinal, err)
		}
		hash, err := validateStepRow(claim.Job.PlanSHA256, step, row)
		if err != nil {
			return nil, fmt.Errorf("step %d: %w", ordinal, err)
		}
		if err := validateStepProgress(step.OperationID, row); err != nil {
			return nil, fmt.Errorf("step %d: %w", ordinal, err)
		}
		validated[ordinal] = ValidatedStep{
			contractVersion: schemav1.PlanContractVersion,
			value:           step, arguments: arguments, hash: hash,
		}
	}
	current := validated[claim.Job.NextStepOrdinal]
	if current.value.ExecutorKind != schemav1.MigratorExecutor ||
		current.hash != claim.Lease.StepSHA256 || current.value.Ordinal != claim.Lease.StepOrdinal {
		return nil, fmt.Errorf("claimed step is not the exact migrator step")
	}
	if claim.Steps[claim.Job.NextStepOrdinal].Generation != claim.Lease.StepGeneration {
		return nil, fmt.Errorf("claimed step generation differs")
	}
	if claim.Steps[claim.Job.NextStepOrdinal].Attempt != claim.Lease.StepAttempt {
		return nil, fmt.Errorf("claimed step attempt differs")
	}
	return validated, nil
}

func validateStepProgress(operation schemav1.OperationID, row StepRecord) error {
	if operation != schemav1.BackfillFieldBatch {
		if row.CheckpointRecordID != nil || row.ProcessedRows != 0 {
			return fmt.Errorf("non-backfill step carries progress")
		}
		return nil
	}
	if row.ProcessedRows < 0 || (row.ProcessedRows == 0) != (row.CheckpointRecordID == nil) {
		return fmt.Errorf("backfill progress shape is invalid")
	}
	if row.CheckpointRecordID != nil {
		if _, err := canonicalUUID(*row.CheckpointRecordID); err != nil {
			return fmt.Errorf("backfill checkpoint is invalid")
		}
	}
	return nil
}
