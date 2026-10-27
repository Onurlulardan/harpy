-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_system_resources_apply_finalized_switch_v1(
    p_operation_id uuid, p_run_id uuid, p_acceptance_digest text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_group platform.cutover_groups%ROWTYPE;
    v_members jsonb;
    v_locked_count integer;
    v_changed_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_cutover_finalizer' OR
       p_acceptance_digest IS NULL OR
       p_acceptance_digest !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='P5670',
            MESSAGE='IAM system-resource finalized switch authority denied';
    END IF;

    SELECT operation.* INTO STRICT v_operation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
    FOR UPDATE;
    IF v_operation.state IS DISTINCT FROM 'completed' OR
       v_operation.scope_kind IS DISTINCT FROM 'tenant' OR
       v_operation.group_id IS DISTINCT FROM 'iam_system_resources' OR
       v_operation.domain IS DISTINCT FROM 'iam' THEN
        RAISE EXCEPTION USING ERRCODE='P5671',
            MESSAGE='IAM system-resource finalized operation differs';
    END IF;

    SELECT cutover_group.* INTO STRICT v_group
    FROM platform.cutover_groups AS cutover_group
    WHERE cutover_group.scope_kind=v_operation.scope_kind
      AND cutover_group.scope_id=v_operation.scope_id
      AND cutover_group.group_id=v_operation.group_id
    FOR UPDATE;
    IF v_group.state IS DISTINCT FROM 'quiescing' OR
       v_group.writer IS DISTINCT FROM 'none' OR
       v_group.generation IS DISTINCT FROM v_operation.cutover_generation OR
       v_group.approved_by IS NULL OR
       v_group.approved_by IS NOT DISTINCT FROM v_group.requested_by THEN
        RAISE EXCEPTION USING ERRCODE='P5672',
            MESSAGE='IAM system-resource quiescing group differs';
    END IF;

    SELECT pg_catalog.jsonb_agg(member.domain ORDER BY member.ordinal)
    INTO v_members
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind=v_group.scope_kind AND member.scope_id=v_group.scope_id
      AND member.group_id=v_group.group_id;
    IF v_members IS DISTINCT FROM '["iam","system_resource_layout"]'::jsonb THEN
        RAISE EXCEPTION USING ERRCODE='P5673',
            MESSAGE='IAM system-resource exact members differ';
    END IF;

    PERFORM 1
    FROM platform.cutover_group_members AS member
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=member.scope_kind AND domain_row.scope_id=member.scope_id
     AND domain_row.domain=member.domain
    WHERE member.scope_kind=v_group.scope_kind AND member.scope_id=v_group.scope_id
      AND member.group_id=v_group.group_id
    ORDER BY member.ordinal
    FOR UPDATE OF domain_row;
    GET DIAGNOSTICS v_locked_count = ROW_COUNT;
    IF v_locked_count IS DISTINCT FROM 2 OR EXISTS (
        SELECT 1 FROM platform.cutover_group_members AS member
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind=member.scope_kind AND domain_row.scope_id=member.scope_id
         AND domain_row.domain=member.domain
        WHERE member.scope_kind=v_group.scope_kind AND member.scope_id=v_group.scope_id
          AND member.group_id=v_group.group_id AND
          (domain_row.state IS DISTINCT FROM v_group.state OR
           domain_row.writer IS DISTINCT FROM v_group.writer OR
           domain_row.generation IS DISTINCT FROM v_group.generation OR
           domain_row.requested_by IS DISTINCT FROM v_group.requested_by OR
           domain_row.approved_by IS DISTINCT FROM v_group.approved_by OR
           domain_row.reason IS DISTINCT FROM v_group.reason OR
           domain_row.change_reference IS DISTINCT FROM v_group.change_reference)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P5673',
            MESSAGE='IAM system-resource member authority differs';
    END IF;

    v_changed_at:=pg_catalog.clock_timestamp();
    PERFORM 1
    FROM platform.cutover_group_members AS member
    JOIN platform.async_v1_adoption_runs AS adoption
      ON adoption.scope_kind=member.scope_kind AND adoption.scope_id=member.scope_id
     AND adoption.group_id=member.group_id AND adoption.domain=member.domain
     AND adoption.source_generation=v_group.generation
     AND adoption.target_generation=v_group.generation+1
     AND adoption.source_group_generation=v_group.generation
     AND adoption.target_group_generation=v_group.generation+1
    JOIN platform.async_v1_zero_inventory_receipts AS zero_receipt
      ON zero_receipt.run_id=adoption.run_id
    WHERE member.scope_kind=v_group.scope_kind AND member.scope_id=v_group.scope_id
      AND member.group_id=v_group.group_id AND adoption.state='accepted'
      AND adoption.item_count=0 AND adoption.accepted_at<=v_changed_at
      AND adoption.acceptance_sha256=p_acceptance_digest
      AND zero_receipt.inventory_sha256=adoption.inventory_sha256
      AND zero_receipt.queue_count=0 AND zero_receipt.timer_count=0
      AND zero_receipt.future_count=0 AND zero_receipt.running_count=0
    ORDER BY member.ordinal
    FOR SHARE OF adoption,zero_receipt;
    GET DIAGNOSTICS v_locked_count = ROW_COUNT;
    IF v_locked_count IS DISTINCT FROM 2 THEN
        RAISE EXCEPTION USING ERRCODE='P5674',
            MESSAGE='IAM system-resource switch requires two accepted zero inventories';
    END IF;

    WITH targets AS (
        SELECT 'group'::text AS target_kind,NULL::text AS domain,
            v_group.requested_by AS requested_by,v_group.approved_by AS approved_by,
            v_group.reason AS reason,v_group.change_reference AS change_reference
        UNION ALL
        SELECT 'domain',domain_row.domain,domain_row.requested_by,domain_row.approved_by,
            domain_row.reason,domain_row.change_reference
        FROM platform.cutover_group_members AS member
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind=member.scope_kind AND domain_row.scope_id=member.scope_id
         AND domain_row.domain=member.domain
        WHERE member.scope_kind=v_group.scope_kind AND member.scope_id=v_group.scope_id
          AND member.group_id=v_group.group_id
    )
    INSERT INTO platform.cutover_transition_events (
        event_id,request_event_id,target_kind,event_kind,scope_kind,scope_id,
        group_id,domain,previous_state,new_state,previous_writer,new_writer,
        previous_generation,new_generation,actor_id,requested_by,approved_by,
        reason,change_reference,acceptance_digest,occurred_at
    )
    SELECT pg_catalog.gen_random_uuid(),request.event_id,target.target_kind,'switch',
        v_group.scope_kind,v_group.scope_id,v_group.group_id,target.domain,
        'quiescing','v2','none','v2',v_group.generation,v_group.generation+1,
        target.approved_by,target.requested_by,target.approved_by,target.reason,
        target.change_reference,p_acceptance_digest,v_changed_at
    FROM targets AS target
    JOIN LATERAL (
        SELECT event.event_id FROM platform.cutover_transition_events AS event
        WHERE event.target_kind=target.target_kind AND event.event_kind='request'
          AND event.scope_kind=v_group.scope_kind AND event.scope_id=v_group.scope_id
          AND event.group_id=v_group.group_id
          AND event.domain IS NOT DISTINCT FROM target.domain
          AND event.new_state='quiescing' AND event.new_generation=v_group.generation
          AND event.requested_by=target.requested_by
          AND event.change_reference=target.change_reference
        ORDER BY event.occurred_at DESC,event.event_id DESC LIMIT 1
    ) AS request ON true;
    GET DIAGNOSTICS v_locked_count = ROW_COUNT;
    IF v_locked_count IS DISTINCT FROM 3 THEN
        RAISE EXCEPTION USING ERRCODE='P5675',
            MESSAGE='IAM system-resource switch request evidence differs';
    END IF;

    UPDATE platform.domain_cutovers AS domain_row
    SET state='v2',writer='v2',generation=domain_row.generation+1,
        acceptance_digest=p_acceptance_digest,changed_at=v_changed_at
    WHERE domain_row.scope_kind=v_group.scope_kind AND domain_row.scope_id=v_group.scope_id
      AND domain_row.domain IN ('iam','system_resource_layout')
      AND domain_row.state='quiescing'
      AND domain_row.writer='none' AND domain_row.generation=v_group.generation;
    GET DIAGNOSTICS v_locked_count = ROW_COUNT;
    IF v_locked_count IS DISTINCT FROM 2 THEN
        RAISE EXCEPTION USING ERRCODE='P5676',
            MESSAGE='IAM system-resource domain switch lost its lock';
    END IF;
    UPDATE platform.cutover_groups AS cutover_group
    SET state='v2',writer='v2',generation=cutover_group.generation+1,
        acceptance_digest=p_acceptance_digest,changed_at=v_changed_at
    WHERE cutover_group.scope_kind=v_group.scope_kind AND cutover_group.scope_id=v_group.scope_id
      AND cutover_group.group_id=v_group.group_id AND cutover_group.state='quiescing'
      AND cutover_group.writer='none' AND cutover_group.generation=v_group.generation;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P5676',
            MESSAGE='IAM system-resource group switch lost its lock';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_system_resources_apply_finalized_switch_v1(
    uuid,uuid,text
) FROM PUBLIC;
