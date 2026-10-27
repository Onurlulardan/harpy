//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
)

func seedGlobalBillingSuccessorFixture(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	index int, operationGeneration int64,
) successorFixture {
	t.Helper()
	base := index*100 + 10
	operation := successorOperation{
		domain: "billing", operationID: successorUUID(base),
		runID: successorUUID(base + 1), leaseID: successorUUID(base + 2),
	}
	fixture := successorFixture{
		scopeKind: "global", scopeID: successorUUID(index * 100),
		groupID: "billing", generation: 6,
		operations: []successorOperation{operation},
	}
	reference := fmt.Sprintf("GLOBAL-BILLING-SUCCESSOR-%d", index)
	publication := fmt.Sprintf("global_billing_%d", index)
	_, err := database.admin.Exec(ctx, `
INSERT INTO platform.cutover_groups (
 scope_kind,scope_id,group_id,state,writer,generation,requested_by,approved_by,
 reason,change_reference
) VALUES ('global',$1,'billing','quiescing','none',6,$2,$3,
 'global Billing successor PG18',$4);
INSERT INTO platform.domain_cutovers (
 scope_kind,scope_id,domain,state,writer,generation,requested_by,approved_by,
 reason,change_reference
) VALUES ('global',$1,'billing','quiescing','none',6,$2,$3,
 'global Billing successor PG18',$4);
INSERT INTO platform.cutover_group_members (
 scope_kind,scope_id,group_id,ordinal,domain
) VALUES ('global',$1,'billing',1,'billing');
INSERT INTO platform.cutover_transition_events (
 event_id,request_event_id,target_kind,event_kind,scope_kind,scope_id,group_id,
 domain,previous_state,new_state,previous_writer,new_writer,previous_generation,
 new_generation,actor_id,requested_by,approved_by,reason,change_reference
) VALUES
 (pg_catalog.uuidv7(),NULL,'group','request','global',$1,'billing',NULL,
  'shadow','quiescing','v1','none',6,6,$2,$2,NULL,
  'global Billing successor PG18',$4),
 (pg_catalog.uuidv7(),NULL,'domain','request','global',$1,'billing','billing',
  'shadow','quiescing','v1','none',6,6,$2,$2,NULL,
  'global Billing successor PG18',$4);
INSERT INTO platform.cutover_operations (
 operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
 group_id,domain,cutover_generation,lease_generation,lease_owner,
 lease_expires_at,last_heartbeat_at,lease_manifest_sha256,plan_sha256,
 transform_manifest_sha256,source_system_identifier,source_database_oid,
 target_system_identifier,target_database_oid,publication_name,slot_name,state
) SELECT $5,$6,$7,$8,$9,'global',$1,'billing','billing',$10,1,$2,
 clock_timestamp()+interval '1 hour',clock_timestamp(),repeat('a',64),
 repeat('b',64),repeat('c',64),control.system_identifier,database_row.oid,
 control.system_identifier,database_row.oid,$11,$11||'_slot','quiescing'
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database_row ON database_row.datname=current_database();`,
		pgx.QueryExecModeSimpleProtocol, fixture.scopeID, successorRequestedBy,
		successorApprovedBy, reference, operation.operationID, operation.runID,
		operation.leaseID, integrationReleaseID, integrationEnvironment,
		operationGeneration, publication)
	if err != nil {
		t.Fatalf("seed global Billing successor: %v", err)
	}
	return fixture
}
