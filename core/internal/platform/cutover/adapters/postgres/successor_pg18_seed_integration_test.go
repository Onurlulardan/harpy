//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	successorRequestedBy = "71000000-0000-4000-8000-000000000001"
	successorApprovedBy  = "71000000-0000-4000-8000-000000000002"
)

type successorOperation struct {
	domain, operationID, runID, leaseID string
}

type successorFixture struct {
	scopeKind, scopeID, groupID string
	generation                  int64
	operations                  []successorOperation
}

func successorUUID(value int) string {
	return fmt.Sprintf("71000000-0000-4000-8000-%012d", value)
}

func seedSuccessorFixture(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	index int, groupID string, domains []string, operationGeneration int64,
) successorFixture {
	t.Helper()
	fixture := successorFixture{
		scopeKind: "tenant", scopeID: successorUUID(index * 100),
		groupID: groupID, generation: 6,
		operations: make([]successorOperation, 0, len(domains)),
	}
	reference := fmt.Sprintf("SUCCESSOR-%d", index)
	_, err := database.admin.Exec(ctx, `
INSERT INTO platform.cutover_groups (
 scope_kind,scope_id,group_id,state,writer,generation,requested_by,approved_by,
 reason,change_reference
) VALUES ('tenant',$1,$2,'quiescing','none',$3,$4,$5,'successor PG18',$6);
INSERT INTO platform.cutover_transition_events (
 event_id,request_event_id,target_kind,event_kind,scope_kind,scope_id,group_id,
 domain,previous_state,new_state,previous_writer,new_writer,previous_generation,
 new_generation,actor_id,requested_by,approved_by,reason,change_reference
) VALUES (pg_catalog.uuidv7(),NULL,'group','request','tenant',$1,$2,NULL,
 'shadow','quiescing','v1','none',$3,$3,$4,$4,NULL,'successor PG18',$6);`,
		pgx.QueryExecModeSimpleProtocol, fixture.scopeID, groupID,
		fixture.generation, successorRequestedBy, successorApprovedBy, reference)
	if err != nil {
		t.Fatalf("seed successor group %s: %v", groupID, err)
	}
	for ordinal, domain := range domains {
		base := index*100 + 10 + ordinal*10
		operation := successorOperation{
			domain: domain, operationID: successorUUID(base),
			runID: successorUUID(base + 1), leaseID: successorUUID(base + 2),
		}
		fixture.operations = append(fixture.operations, operation)
		publication := fmt.Sprintf("successor_%d_%d", index, ordinal)
		slot := publication + "_slot"
		_, err = database.admin.Exec(ctx, `
INSERT INTO platform.domain_cutovers (
 scope_kind,scope_id,domain,state,writer,generation,requested_by,approved_by,
 reason,change_reference
) VALUES ('tenant',$1,$2,'quiescing','none',$3,$4,$5,'successor PG18',$6);
INSERT INTO platform.cutover_group_members (
 scope_kind,scope_id,group_id,ordinal,domain
) SELECT 'tenant',$1,$7,$8,$2 WHERE $7='billing_forms';
INSERT INTO platform.cutover_transition_events (
 event_id,request_event_id,target_kind,event_kind,scope_kind,scope_id,group_id,
 domain,previous_state,new_state,previous_writer,new_writer,previous_generation,
 new_generation,actor_id,requested_by,approved_by,reason,change_reference
) VALUES (pg_catalog.uuidv7(),NULL,'domain','request','tenant',$1,$7,$2,
 'shadow','quiescing','v1','none',$3,$3,$4,$4,NULL,'successor PG18',$6);
INSERT INTO platform.cutover_operations (
 operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
 group_id,domain,cutover_generation,lease_generation,lease_owner,
 lease_expires_at,last_heartbeat_at,lease_manifest_sha256,plan_sha256,
 transform_manifest_sha256,source_system_identifier,source_database_oid,
 target_system_identifier,target_database_oid,publication_name,slot_name,state
) SELECT $9,$10,$11,$12,$13,'tenant',$1,$7,$2,$14,1,$4,
 clock_timestamp()+interval '1 hour',clock_timestamp(),repeat('a',64),
 repeat('b',64),repeat('c',64),control.system_identifier,database_row.oid,
 control.system_identifier,database_row.oid,$15,$16,'quiescing'
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database_row ON database_row.datname=current_database();`,
			pgx.QueryExecModeSimpleProtocol, fixture.scopeID, domain,
			fixture.generation, successorRequestedBy, successorApprovedBy, reference,
			groupID, ordinal+1, operation.operationID, operation.runID,
			operation.leaseID, integrationReleaseID, integrationEnvironment,
			operationGeneration, publication, slot)
		if err != nil {
			t.Fatalf("seed successor domain %s: %v", domain, err)
		}
	}
	return fixture
}

func seedSuccessorZeroInventory(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	fixture successorFixture,
) {
	t.Helper()
	for _, operation := range fixture.operations {
		_, err := database.admin.Exec(ctx, `WITH stamp AS (
 SELECT pg_catalog.clock_timestamp() AS now
),adoption AS (
 INSERT INTO platform.async_v1_adoption_runs (
  run_id,scope_kind,scope_id,domain,group_id,source_generation,target_generation,
  source_group_generation,target_group_generation,inventory_contract_version,
  inventory_revision,inventory_sha256,component_manifest_sha256,
  expected_enqueue_fences,expected_scheduler_fences,expected_consumer_fences,
  item_count,state,captured_at,created_at,accepted_at,acceptance_sha256
 ) SELECT pg_catalog.uuidv7(),$1,$2,$3,$4,$5::bigint,$5::bigint+1,
  $5::bigint,$5::bigint+1,
  'async-adoption-inventory.v1',1,repeat('d',64),repeat('e',64),1,1,1,0,
  'accepted',now-interval '2 seconds',now-interval '1 second',now,$6 FROM stamp
 RETURNING run_id,inventory_sha256
)
INSERT INTO platform.async_v1_zero_inventory_receipts (
 run_id,inventory_contract_version,inventory_sha256,queue_count,timer_count,
 future_count,running_count,observed_at,receipt_sha256
) SELECT run_id,'async-adoption-zero-inventory.v1',inventory_sha256,0,0,0,0,
 clock_timestamp(),repeat('f',64) FROM adoption;`, fixture.scopeKind,
			fixture.scopeID, operation.domain, fixture.groupID, fixture.generation,
			successorAcceptanceDigest)
		if err != nil {
			t.Fatalf("seed successor zero inventory for %s: %v", operation.domain, err)
		}
	}
}
