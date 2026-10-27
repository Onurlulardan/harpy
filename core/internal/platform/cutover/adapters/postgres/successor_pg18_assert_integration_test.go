//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func allowSuccessorReceipt(
	t *testing.T, ctx context.Context, admin *pgx.Conn, operation successorOperation,
) {
	t.Helper()
	_, err := admin.Exec(ctx, `
INSERT INTO cutover_authority.successor_test_receipt_permissions
 (operation_id,run_id,lease_id,digest) VALUES ($1,$2,$3,$4)`,
		operation.operationID, operation.runID, operation.leaseID,
		successorAcceptanceDigest)
	if err != nil {
		t.Fatalf("allow successor receipt for %s: %v", operation.domain, err)
	}
}

func finalizeSuccessor(
	ctx context.Context, finalizer *pgx.Conn, operation successorOperation,
) error {
	_, err := finalizer.Exec(ctx, `
SELECT cutover_authority.successor_test_finalize_v1($1,$2,$3,$4)`,
		operation.operationID, operation.runID, operation.leaseID,
		successorAcceptanceDigest)
	return err
}

func assertSuccessorRoute(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	fixture successorFixture, state, writer string, generation int64,
	callCount, switchCount int,
) {
	t.Helper()
	var actualState, actualWriter, digest string
	var actualGeneration int64
	var domains, calls, switches int
	err := admin.QueryRow(ctx, `
SELECT group_row.state,group_row.writer,group_row.generation,
       COALESCE(group_row.acceptance_digest,''),
       (SELECT count(*) FROM platform.domain_cutovers AS domain_row
        JOIN platform.cutover_group_members AS member
          ON member.scope_kind=domain_row.scope_kind
         AND member.scope_id=domain_row.scope_id AND member.domain=domain_row.domain
        WHERE member.scope_kind=group_row.scope_kind
          AND member.scope_id=group_row.scope_id AND member.group_id=group_row.group_id
          AND domain_row.state=$4 AND domain_row.writer=$5
          AND domain_row.generation=$6),
       (SELECT count(*) FROM platform.successor_authority_test_calls AS call
        WHERE call.operation_id=ANY($7::uuid[])),
       (SELECT count(*) FROM platform.cutover_transition_events AS event
        WHERE event.scope_kind=group_row.scope_kind
          AND event.scope_id=group_row.scope_id AND event.group_id=group_row.group_id
          AND event.event_kind='switch')
FROM platform.cutover_groups AS group_row
WHERE group_row.scope_kind=$1 AND group_row.scope_id=$2
  AND group_row.group_id=$3`, fixture.scopeKind, fixture.scopeID,
		fixture.groupID, state, writer, generation,
		successorOperationIDs(fixture)).Scan(
		&actualState, &actualWriter, &actualGeneration, &digest,
		&domains, &calls, &switches)
	if err != nil {
		t.Fatalf("read successor route %s: %v", fixture.groupID, err)
	}
	if actualState != state || actualWriter != writer ||
		actualGeneration != generation || domains != len(fixture.operations) ||
		calls != callCount || switches != switchCount {
		t.Fatalf("successor route %s = %s/%s/%d domains=%d calls=%d switches=%d",
			fixture.groupID, actualState, actualWriter, actualGeneration,
			domains, calls, switches)
	}
	if state == "v2" && digest != successorAcceptanceDigest {
		t.Fatalf("successor route %s digest = %q", fixture.groupID, digest)
	}
	if state != "v2" && digest != "" {
		t.Fatalf("unswitched successor route %s digest = %q", fixture.groupID, digest)
	}
}

func successorOperationIDs(fixture successorFixture) []string {
	ids := make([]string, 0, len(fixture.operations))
	for _, operation := range fixture.operations {
		ids = append(ids, operation.operationID)
	}
	return ids
}

func assertSuccessorOperationState(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	operation successorOperation, state string, receiptCount int,
) {
	t.Helper()
	var actual string
	var receipts int
	err := admin.QueryRow(ctx, `SELECT operation.state,
 (SELECT count(*) FROM cutover_authority.finalization_receipts AS receipt
  WHERE receipt.operation_id=operation.operation_id AND receipt.run_id=operation.run_id)
FROM platform.cutover_operations AS operation
WHERE operation.operation_id=$1 AND operation.run_id=$2`,
		operation.operationID, operation.runID).Scan(&actual, &receipts)
	if err != nil || actual != state || receipts != receiptCount {
		t.Fatalf("successor operation %s = %s receipts=%d err=%v",
			operation.domain, actual, receipts, err)
	}
}

func assertSuccessorAtomicXID(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	fixture successorFixture, triggering successorOperation,
) {
	t.Helper()
	var total, distinct int
	err := admin.QueryRow(ctx, `SELECT count(*),count(DISTINCT xid) FROM (
 SELECT group_row.xmin::text AS xid FROM platform.cutover_groups AS group_row
  WHERE group_row.scope_kind=$1 AND group_row.scope_id=$2
    AND group_row.group_id=$3
 UNION ALL SELECT domain_row.xmin::text FROM platform.domain_cutovers AS domain_row
  JOIN platform.cutover_group_members AS member
    ON member.scope_kind=domain_row.scope_kind AND member.scope_id=domain_row.scope_id
   AND member.domain=domain_row.domain
  WHERE member.scope_kind=$1 AND member.scope_id=$2 AND member.group_id=$3
 UNION ALL SELECT call.call_xid::text FROM platform.successor_authority_test_calls AS call
  WHERE call.operation_id=ANY($4::uuid[])
 UNION ALL SELECT event.event_xid::text FROM platform.cutover_transition_events AS event
  WHERE event.scope_kind=$1 AND event.scope_id=$2 AND event.group_id=$3
    AND event.event_kind='switch'
 UNION ALL SELECT receipt.xmin::text FROM cutover_authority.finalization_receipts AS receipt
  WHERE receipt.operation_id=$5 AND receipt.run_id=$6
) AS evidence(xid)`, fixture.scopeKind, fixture.scopeID, fixture.groupID,
		successorOperationIDs(fixture), triggering.operationID, triggering.runID).
		Scan(&total, &distinct)
	wantTotal := 2*len(fixture.operations) + 4
	if err != nil || total != wantTotal || distinct != 1 {
		t.Fatalf("successor atomic XID %s total=%d/%d distinct=%d err=%v",
			fixture.groupID, total, wantTotal, distinct, err)
	}
}
