-- +goose Up
-- Migration stream version: 20260729100011.
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_finalize_successor_group_v1(p_operation_id uuid,p_run_id uuid,p_companion_operation_id uuid,p_companion_run_id uuid,p_acceptance_digest text)
RETURNS uuid LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE; v_companion platform.cutover_operations%ROWTYPE; v_group platform.cutover_groups%ROWTYPE;
    v_members jsonb; v_expected jsonb; v_changed_at timestamptz;
    v_count integer; v_authority_id uuid;
    v_billing_operation_id uuid; v_billing_run_id uuid; v_forms_operation_id uuid; v_forms_run_id uuid;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR SESSION_USER IS DISTINCT FROM 'harpy_cutover_finalizer'
       OR p_acceptance_digest !~ '^[0-9a-f]{64}$'
       OR (p_companion_operation_id IS NULL) IS DISTINCT FROM
          (p_companion_run_id IS NULL) THEN
        RAISE EXCEPTION USING ERRCODE='P5734', MESSAGE='cutover successor group authority is denied';
    END IF;
    SELECT operation.* INTO STRICT v_operation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
      AND operation.state='completed' AND (
        (operation.scope_kind='tenant' AND operation.group_id IN ('analytics','billing_forms','workflow')
         AND operation.domain IN ('analytics','billing','forms','workflow')) OR
        (operation.scope_kind='global' AND operation.group_id='billing' AND operation.domain='billing'))
    FOR SHARE;
    IF v_operation.group_id='billing_forms' THEN
        SELECT operation.* INTO STRICT v_companion
        FROM platform.cutover_operations AS operation
        WHERE operation.operation_id=p_companion_operation_id
          AND operation.run_id=p_companion_run_id AND operation.state='completed'
          AND operation.scope_kind=v_operation.scope_kind
          AND operation.scope_id=v_operation.scope_id
          AND operation.group_id=v_operation.group_id
          AND operation.cutover_generation=v_operation.cutover_generation
          AND operation.domain=CASE v_operation.domain
              WHEN 'billing' THEN 'forms' ELSE 'billing' END
        FOR SHARE;
        IF v_operation.domain NOT IN ('billing','forms') THEN
            RAISE EXCEPTION USING ERRCODE='P5735',
                MESSAGE='billing forms successor operations differ';
        END IF;
        v_expected:='["billing","forms"]'::jsonb;
        IF v_operation.domain='billing' THEN
            v_billing_operation_id:=p_operation_id; v_billing_run_id:=p_run_id;
            v_forms_operation_id:=p_companion_operation_id;
            v_forms_run_id:=p_companion_run_id;
        ELSE
            v_billing_operation_id:=p_companion_operation_id;
            v_billing_run_id:=p_companion_run_id;
            v_forms_operation_id:=p_operation_id; v_forms_run_id:=p_run_id;
        END IF;
    ELSE
        IF p_companion_operation_id IS NOT NULL
           OR v_operation.group_id IS DISTINCT FROM v_operation.domain THEN
            RAISE EXCEPTION USING ERRCODE='P5735',
                MESSAGE='singleton successor operation differs';
        END IF;
        v_expected:=pg_catalog.jsonb_build_array(v_operation.domain);
    END IF;
    SELECT group_row.* INTO STRICT v_group
    FROM platform.cutover_groups AS group_row
    WHERE group_row.scope_kind=v_operation.scope_kind
      AND group_row.scope_id=v_operation.scope_id
      AND group_row.group_id=v_operation.group_id
      AND group_row.state='quiescing' AND group_row.writer='none'
      AND group_row.generation=v_operation.cutover_generation
      AND group_row.approved_by IS NOT NULL
      AND group_row.approved_by IS DISTINCT FROM group_row.requested_by
    FOR UPDATE;
    SELECT pg_catalog.jsonb_agg(member.domain ORDER BY member.ordinal)
    INTO v_members FROM platform.cutover_group_members AS member
    WHERE member.scope_kind=v_group.scope_kind AND member.scope_id=v_group.scope_id
      AND member.group_id=v_group.group_id;
    IF v_members IS DISTINCT FROM v_expected THEN
        RAISE EXCEPTION USING ERRCODE='P5736',
            MESSAGE='cutover successor group members differ';
    END IF;
    v_changed_at:=pg_catalog.clock_timestamp();
    PERFORM 1 FROM platform.cutover_group_members AS member
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=member.scope_kind AND domain_row.scope_id=member.scope_id
     AND domain_row.domain=member.domain
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
      AND member.group_id=v_group.group_id
      AND domain_row.state='quiescing' AND domain_row.writer='none'
      AND domain_row.generation=v_group.generation
      AND domain_row.requested_by=v_group.requested_by
      AND domain_row.approved_by IS NOT DISTINCT FROM v_group.approved_by
      AND domain_row.reason=v_group.reason
      AND domain_row.change_reference=v_group.change_reference
      AND adoption.state='accepted' AND adoption.item_count=0
      AND adoption.accepted_at<=v_changed_at
      AND adoption.acceptance_sha256=p_acceptance_digest
      AND zero_receipt.inventory_sha256=adoption.inventory_sha256
      AND zero_receipt.queue_count=0 AND zero_receipt.timer_count=0
      AND zero_receipt.future_count=0 AND zero_receipt.running_count=0
    ORDER BY member.ordinal
    FOR UPDATE OF domain_row FOR SHARE OF adoption,zero_receipt;
    GET DIAGNOSTICS v_count=ROW_COUNT;
    IF v_count IS DISTINCT FROM pg_catalog.jsonb_array_length(v_expected) THEN
        RAISE EXCEPTION USING ERRCODE='P5737',
            MESSAGE='cutover successor zero inventory differs';
    END IF;
    WITH targets AS (
        SELECT 'group'::text AS target_kind,NULL::text AS domain,
            v_group.requested_by AS requested_by,v_group.approved_by AS approved_by,
            v_group.reason AS reason,v_group.change_reference AS change_reference
        UNION ALL
        SELECT 'domain',domain_row.domain,domain_row.requested_by,
            domain_row.approved_by,domain_row.reason,domain_row.change_reference
        FROM platform.cutover_group_members AS member
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind=member.scope_kind
         AND domain_row.scope_id=member.scope_id AND domain_row.domain=member.domain
        WHERE member.scope_kind=v_group.scope_kind AND member.scope_id=v_group.scope_id
          AND member.group_id=v_group.group_id
    )
    INSERT INTO platform.cutover_transition_events (
        event_id,request_event_id,target_kind,event_kind,scope_kind,scope_id,
        group_id,domain,previous_state,new_state,previous_writer,new_writer,
        previous_generation,new_generation,actor_id,requested_by,approved_by,
        reason,change_reference,acceptance_digest,occurred_at
    )
    SELECT pg_catalog.uuidv7(),request.event_id,target.target_kind,'switch',
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
    GET DIAGNOSTICS v_count=ROW_COUNT;
    IF v_count IS DISTINCT FROM pg_catalog.jsonb_array_length(v_expected)+1 THEN
        RAISE EXCEPTION USING ERRCODE='P5738',
            MESSAGE='cutover successor transition evidence differs';
    END IF;
    UPDATE platform.domain_cutovers AS domain_row
    SET state='v2',writer='v2',generation=domain_row.generation+1,
        acceptance_digest=p_acceptance_digest,changed_at=v_changed_at
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind=v_group.scope_kind AND member.scope_id=v_group.scope_id
      AND member.group_id=v_group.group_id
      AND domain_row.scope_kind=member.scope_kind
      AND domain_row.scope_id=member.scope_id AND domain_row.domain=member.domain
      AND domain_row.state='quiescing' AND domain_row.writer='none'
      AND domain_row.generation=v_group.generation;
    GET DIAGNOSTICS v_count=ROW_COUNT;
    IF v_count IS DISTINCT FROM pg_catalog.jsonb_array_length(v_expected) THEN
        RAISE EXCEPTION USING ERRCODE='P5739',
            MESSAGE='cutover successor domain switch lost its lock';
    END IF;
    UPDATE platform.cutover_groups AS group_row
    SET state='v2',writer='v2',generation=group_row.generation+1,
        acceptance_digest=p_acceptance_digest,changed_at=v_changed_at
    WHERE group_row.scope_kind=v_group.scope_kind AND group_row.scope_id=v_group.scope_id
      AND group_row.group_id=v_group.group_id AND group_row.state='quiescing'
      AND group_row.writer='none' AND group_row.generation=v_group.generation;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P5739',
            MESSAGE='cutover successor group switch lost its lock';
    END IF;
    IF v_group.group_id='analytics' THEN
        SELECT platform.analytics_finalize_authority_successor_v1(
            p_operation_id,p_run_id,p_acceptance_digest) INTO STRICT v_authority_id;
    ELSIF v_group.group_id='workflow' THEN
        SELECT platform.workflow_finalize_authority_successor_v1(
            p_operation_id,p_run_id,p_acceptance_digest) INTO STRICT v_authority_id;
    ELSIF v_group.group_id='billing_forms' THEN
        SELECT platform.billing_finalize_authority_successor_v1(v_billing_operation_id,
            v_billing_run_id,v_forms_operation_id,v_forms_run_id,p_acceptance_digest) INTO STRICT v_authority_id;
    ELSE
        SELECT platform.billing_finalize_authority_successor_v1(
            p_operation_id,p_run_id,NULL,NULL,p_acceptance_digest) INTO STRICT v_authority_id;
    END IF;
    RETURN v_authority_id;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.cutover_finalize_successor_group_v1(uuid,uuid,uuid,uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.cutover_finalize_successor_group_v1(uuid,uuid,uuid,uuid,text) TO harpy_cutover_authority_owner;
