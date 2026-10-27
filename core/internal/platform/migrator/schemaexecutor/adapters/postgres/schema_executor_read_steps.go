package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func schemaExecutorReadSteps(
	ctx context.Context,
	transaction pgx.Tx,
	jobID string,
) ([]schemaexecutor.StepRecord, error) {
	rows, err := transaction.Query(ctx, schemaExecutorReadJobStepsSQL, jobID)
	if err != nil {
		return nil, fmt.Errorf("read schema job steps: %w", err)
	}
	defer rows.Close()
	var result []schemaexecutor.StepRecord
	for rows.Next() {
		var step schemaexecutor.StepRecord
		var identifiers, arguments string
		if err := rows.Scan(
			&step.Ordinal, &step.Generation, &step.Attempt,
			&step.OperationID, &step.TransactionClass,
			&step.ExecutorKind, &identifiers, &arguments, &step.StepSHA256,
			&step.CheckpointRecordID, &step.ProcessedRows,
		); err != nil {
			return nil, fmt.Errorf("scan schema job step: %w", err)
		}
		step.IdentifierMapping = []byte(identifiers)
		step.Arguments = []byte(arguments)
		result = append(result, step)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schema job steps: %w", err)
	}
	return result, nil
}
