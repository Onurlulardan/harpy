//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	iamSystemResourcesApproverID = "32000000-0000-4000-8000-000000000001"
	iamSystemResourcesOtherScope = "32000000-0000-4000-8000-000000000002"
)

func prepareIAMSystemResourcesQuiescing(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	scopeID, acceptanceSHA256 string,
) {
	t.Helper()
	var requests, approvals, quiesces, domains, groups int
	err := database.admin.QueryRow(ctx, `WITH stamp AS (
    SELECT pg_catalog.clock_timestamp() AS now
), targets AS (
    SELECT 'group'::text AS target_kind,NULL::text AS domain
    UNION ALL SELECT 'domain','iam'
    UNION ALL SELECT 'domain','system_resource_layout'
), requests AS (
    INSERT INTO platform.cutover_transition_events (
        event_id,request_event_id,target_kind,event_kind,scope_kind,scope_id,
        group_id,domain,previous_state,new_state,previous_writer,new_writer,
        previous_generation,new_generation,actor_id,requested_by,approved_by,
        reason,change_reference,acceptance_digest,occurred_at
    ) SELECT pg_catalog.gen_random_uuid(),NULL,target.target_kind,'request',
        'tenant',$1,'iam_system_resources',target.domain,'shadow','quiescing',
        'v1','none',1,1,$2,$2,NULL,'IAM system-resource protected switch',
        'IAM-SYSTEM-RESOURCES-FINALIZER',NULL,stamp.now
    FROM targets AS target CROSS JOIN stamp
    RETURNING event_id,target_kind,domain
), approvals AS (
    INSERT INTO platform.cutover_transition_events (
        event_id,request_event_id,target_kind,event_kind,scope_kind,scope_id,
        group_id,domain,previous_state,new_state,previous_writer,new_writer,
        previous_generation,new_generation,actor_id,requested_by,approved_by,
        reason,change_reference,acceptance_digest,occurred_at
    ) SELECT pg_catalog.gen_random_uuid(),request.event_id,target.target_kind,'approval',
        'tenant',$1,'iam_system_resources',target.domain,'shadow','quiescing',
        'v1','none',1,1,$3,$2,$3,'IAM system-resource protected switch',
        'IAM-SYSTEM-RESOURCES-FINALIZER',$4,stamp.now
    FROM targets AS target JOIN requests AS request
      ON request.target_kind=target.target_kind
     AND request.domain IS NOT DISTINCT FROM target.domain
    CROSS JOIN stamp RETURNING 1
), quiesces AS (
    INSERT INTO platform.cutover_transition_events (
        event_id,request_event_id,target_kind,event_kind,scope_kind,scope_id,
        group_id,domain,previous_state,new_state,previous_writer,new_writer,
        previous_generation,new_generation,actor_id,requested_by,approved_by,
        reason,change_reference,acceptance_digest,occurred_at
    ) SELECT pg_catalog.gen_random_uuid(),request.event_id,target.target_kind,'quiesce',
        'tenant',$1,'iam_system_resources',target.domain,'shadow','quiescing',
        'v1','none',1,1,$3,$2,$3,'IAM system-resource protected switch',
        'IAM-SYSTEM-RESOURCES-FINALIZER',$4,stamp.now
    FROM targets AS target JOIN requests AS request
      ON request.target_kind=target.target_kind
     AND request.domain IS NOT DISTINCT FROM target.domain
    CROSS JOIN stamp RETURNING 1
), domain_updates AS (
    UPDATE platform.domain_cutovers AS domain_row
    SET state='quiescing',writer='none',approved_by=$3,
        reason='IAM system-resource protected switch',
        change_reference='IAM-SYSTEM-RESOURCES-FINALIZER',
        acceptance_digest=$4,changed_at=stamp.now
    FROM stamp WHERE domain_row.scope_kind='tenant' AND domain_row.scope_id=$1
      AND domain_row.domain IN ('iam','system_resource_layout')
      AND domain_row.state='shadow' AND domain_row.writer='v1'
    RETURNING 1
), group_update AS (
    UPDATE platform.cutover_groups AS cutover_group
    SET state='quiescing',writer='none',approved_by=$3,
        reason='IAM system-resource protected switch',
        change_reference='IAM-SYSTEM-RESOURCES-FINALIZER',
        acceptance_digest=$4,changed_at=stamp.now
    FROM stamp WHERE cutover_group.scope_kind='tenant' AND cutover_group.scope_id=$1
      AND cutover_group.group_id='iam_system_resources'
      AND cutover_group.state='shadow' AND cutover_group.writer='v1'
    RETURNING 1
)
SELECT (SELECT count(*) FROM requests),(SELECT count(*) FROM approvals),
    (SELECT count(*) FROM quiesces),(SELECT count(*) FROM domain_updates),
    (SELECT count(*) FROM group_update)`, scopeID, integrationOwnerID,
		iamSystemResourcesApproverID, acceptanceSHA256).Scan(
		&requests, &approvals, &quiesces, &domains, &groups,
	)
	if err != nil || requests != 3 || approvals != 3 || quiesces != 3 ||
		domains != 2 || groups != 1 {
		t.Fatalf("prepare IAM system-resource quiescence = %d/%d/%d/%d/%d err=%v",
			requests, approvals, quiesces, domains, groups, err)
	}
}

func seedIsolatedIAMSystemResourcesGroup(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `
INSERT INTO platform.domain_cutovers (
    scope_kind,scope_id,domain,state,writer,generation,requested_by,reason,change_reference
) VALUES
('tenant',$1,'iam','shadow','v1',1,$2,'isolation fixture','IAM-SWITCH-ISOLATION'),
('tenant',$1,'system_resource_layout','shadow','v1',1,$2,'isolation fixture','IAM-SWITCH-ISOLATION');
INSERT INTO platform.cutover_groups (
    scope_kind,scope_id,group_id,state,writer,generation,requested_by,reason,change_reference
) VALUES ('tenant',$1,'iam_system_resources','shadow','v1',1,$2,
    'isolation fixture','IAM-SWITCH-ISOLATION');
INSERT INTO platform.cutover_group_members (
    scope_kind,scope_id,group_id,ordinal,domain
) VALUES
('tenant',$1,'iam_system_resources',1,'iam'),
('tenant',$1,'iam_system_resources',2,'system_resource_layout');`,
		pgx.QueryExecModeSimpleProtocol,
		iamSystemResourcesOtherScope, integrationOwnerID)
	if err != nil {
		t.Fatalf("seed isolated IAM system-resource group: %v", err)
	}
}
