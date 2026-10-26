//go:build integration

package postgres

import (
	"context"
	"testing"
)

func assertMissingLayoutZeroInventoryBlocksSwitch(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) {
	t.Helper()
	connection := connectCutoverRole(
		t, ctx, database.dsn, "harpy_cutover_finalizer",
	)
	_, err := callFinalization(ctx, connection, evidence)
	_ = connection.Close(ctx)
	assertFinalizationError(
		t, err, "P5674", "requires two accepted zero inventories",
	)
	assertIAMSystemResourcesPreSwitch(t, ctx, database, evidence)
}

func assertIAMSystemResourcesFinalizationRollback(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) {
	t.Helper()
	connection := connectCutoverRole(
		t, ctx, database.dsn, "harpy_cutover_finalizer",
	)
	defer connection.Close(ctx)
	transaction, err := connection.Begin(ctx)
	if err != nil {
		t.Fatalf("begin IAM system-resource finalization rollback: %v", err)
	}
	defer transaction.Rollback(ctx)
	outcome, err := callFinalization(ctx, transaction, evidence)
	if err != nil || outcome.state != "completed" ||
		outcome.evidenceSHA256 != evidence.terminalSHA256 {
		t.Fatalf("IAM system-resource rollback candidate = %#v err=%v", outcome, err)
	}
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatalf("rollback IAM system-resource finalization: %v", err)
	}
	assertIAMSystemResourcesPreSwitch(t, ctx, database, evidence)
}

func raceIAMSystemResourcesFinalization(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) {
	t.Helper()
	first := connectFinalizerSession(t, ctx, database.dsn)
	second := connectFinalizerSession(t, ctx, database.dsn)
	outcomes := raceFinalization(ctx, first, second, evidence)
	_ = first.Close(ctx)
	_ = second.Close(ctx)
	assertExpectedFinalizationOutcomes(t, outcomes, evidence)
}

func assertIAMSystemResourcesPreSwitch(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) {
	t.Helper()
	var operationState, groupState, writer string
	var generation int64
	var domains, receipts, switches int
	err := database.admin.QueryRow(ctx, `SELECT operation.state,cutover_group.state,
    cutover_group.writer,cutover_group.generation,
    (SELECT count(*) FROM platform.domain_cutovers AS domain_row
     WHERE domain_row.scope_kind=operation.scope_kind
       AND domain_row.scope_id=operation.scope_id
       AND domain_row.domain IN ('iam','system_resource_layout')
       AND domain_row.state='quiescing' AND domain_row.writer='none'
       AND domain_row.generation=1),
    (SELECT count(*) FROM cutover_authority.finalization_receipts AS receipt
     WHERE receipt.operation_id=operation.operation_id AND receipt.run_id=operation.run_id),
    (SELECT count(*) FROM platform.cutover_transition_events AS event
     WHERE event.scope_kind=operation.scope_kind AND event.scope_id=operation.scope_id
       AND event.group_id=operation.group_id AND event.event_kind='switch')
FROM platform.cutover_operations AS operation
JOIN platform.cutover_groups AS cutover_group
  ON cutover_group.scope_kind=operation.scope_kind
 AND cutover_group.scope_id=operation.scope_id AND cutover_group.group_id=operation.group_id
WHERE operation.operation_id=$1 AND operation.run_id=$2`,
		evidence.operationID, evidence.runID).Scan(
		&operationState, &groupState, &writer, &generation,
		&domains, &receipts, &switches,
	)
	if err != nil || operationState != "quiescing" || groupState != "quiescing" ||
		writer != "none" || generation != 1 || domains != 2 || receipts != 0 || switches != 0 {
		t.Fatalf("IAM pre-switch state = %s/%s/%s/%d/%d/%d/%d err=%v",
			operationState, groupState, writer, generation, domains, receipts, switches, err)
	}
}
