//go:build integration

package postgres

import (
	"context"
	"testing"
)

func assertIAMSystemResourcesFinalizedSwitch(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	scopeID string, evidence finalizationEvidence,
) {
	t.Helper()
	var operationState, groupState, writer, acceptance string
	var generation int64
	var domains, switches, eventXIDs, rowXIDs, receipts, isolated int
	var eventXID, receiptXID string
	err := database.admin.QueryRow(ctx, `SELECT
    (SELECT state FROM platform.cutover_operations WHERE operation_id=$1 AND run_id=$2),
    (SELECT state FROM platform.cutover_groups WHERE scope_kind='tenant' AND scope_id=$3
      AND group_id='iam_system_resources'),
    (SELECT writer FROM platform.cutover_groups WHERE scope_kind='tenant' AND scope_id=$3
      AND group_id='iam_system_resources'),
    (SELECT generation FROM platform.cutover_groups WHERE scope_kind='tenant' AND scope_id=$3
      AND group_id='iam_system_resources'),
    (SELECT acceptance_digest FROM platform.cutover_groups WHERE scope_kind='tenant' AND scope_id=$3
      AND group_id='iam_system_resources'),
    (SELECT count(*) FROM platform.domain_cutovers WHERE scope_kind='tenant' AND scope_id=$3
      AND domain IN ('iam','system_resource_layout') AND state='v2' AND writer='v2'
      AND generation=2 AND acceptance_digest=$4),
    (SELECT count(*) FROM platform.cutover_transition_events WHERE scope_kind='tenant'
      AND scope_id=$3 AND group_id='iam_system_resources' AND event_kind='switch'
      AND previous_state='quiescing' AND new_state='v2' AND acceptance_digest=$4),
    (SELECT count(DISTINCT event_xid::text) FROM platform.cutover_transition_events
      WHERE scope_kind='tenant' AND scope_id=$3 AND group_id='iam_system_resources'
        AND event_kind='switch'),
    (SELECT min(event_xid::text) FROM platform.cutover_transition_events
      WHERE scope_kind='tenant' AND scope_id=$3 AND group_id='iam_system_resources'
        AND event_kind='switch'),
    (SELECT xmin::text FROM cutover_authority.finalization_receipts
      WHERE operation_id=$1 AND run_id=$2),
    (SELECT count(DISTINCT xid) FROM (
      SELECT xmin::text AS xid FROM platform.cutover_groups
        WHERE scope_kind='tenant' AND scope_id=$3 AND group_id='iam_system_resources'
      UNION ALL SELECT xmin::text FROM platform.domain_cutovers
        WHERE scope_kind='tenant' AND scope_id=$3 AND domain IN ('iam','system_resource_layout')
      UNION ALL SELECT xmin::text FROM cutover_authority.finalization_receipts
        WHERE operation_id=$1 AND run_id=$2) AS changed_rows),
    (SELECT count(*) FROM cutover_authority.finalization_receipts
      WHERE operation_id=$1 AND run_id=$2 AND terminal_evidence_sha256=$4),
    (SELECT count(*) FROM platform.cutover_groups AS cutover_group
      WHERE cutover_group.scope_kind='tenant' AND cutover_group.scope_id=$5
        AND cutover_group.group_id='iam_system_resources' AND cutover_group.state='shadow'
        AND cutover_group.writer='v1' AND cutover_group.generation=1)
`, evidence.operationID, evidence.runID, scopeID, evidence.terminalSHA256,
		iamSystemResourcesOtherScope).Scan(
		&operationState, &groupState, &writer, &generation, &acceptance,
		&domains, &switches, &eventXIDs, &eventXID, &receiptXID,
		&rowXIDs, &receipts, &isolated,
	)
	if err != nil || operationState != "completed" || groupState != "v2" ||
		writer != "v2" || generation != 2 || acceptance != evidence.terminalSHA256 ||
		domains != 2 || switches != 3 || eventXIDs != 1 || eventXID != receiptXID ||
		rowXIDs != 1 || receipts != 1 || isolated != 1 {
		t.Fatalf("final IAM switch = %s/%s/%s/%d/%s/%d/%d/%d/%s/%s/%d/%d/%d err=%v",
			operationState, groupState, writer, generation, acceptance, domains,
			switches, eventXIDs, eventXID, receiptXID, rowXIDs, receipts, isolated, err)
	}
}
