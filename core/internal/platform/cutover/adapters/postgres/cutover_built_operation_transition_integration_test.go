//go:build integration

package postgres

import (
	"context"
	"testing"
)

func transitionBuiltOperationToQuiescing(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts,
) {
	t.Helper()
	result, err := database.admin.Exec(ctx, `UPDATE platform.cutover_operations AS operation
SET state='quiescing',updated_at=pg_catalog.clock_timestamp()
WHERE operation.operation_id=$1 AND operation.run_id=$2 AND operation.state='copying'
  AND NOT EXISTS (
    SELECT 1 FROM platform.cutover_operation_checkpoint_keys AS planned
    LEFT JOIN platform.cutover_copy_checkpoints AS checkpoint
      ON checkpoint.operation_id=planned.operation_id AND checkpoint.run_id=planned.run_id
     AND checkpoint.relation_ordinal=planned.relation_ordinal
     AND checkpoint.checkpoint_key=planned.checkpoint_key
    WHERE planned.operation_id=operation.operation_id AND planned.run_id=operation.run_id
      AND (checkpoint.operation_id IS NULL OR checkpoint.state IS DISTINCT FROM 'complete')
  )`, adoption.identity.operationID, adoption.identity.runID)
	if err != nil || result.RowsAffected() != 1 {
		t.Fatalf("transition built operation to quiescing: rows=%d err=%v",
			result.RowsAffected(), err)
	}
}
