-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_queue_evidence_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_id uuid
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_generation bigint;
    v_async_run uuid;
    v_inventory text;
    v_acceptance text;
    v_zero text;
    v_zero_inventory text;
    v_expected_items bigint;
    v_items bigint;
    v_dispositions bigint;
    v_expected_enqueue smallint;
    v_expected_scheduler smallint;
    v_expected_consumer smallint;
    v_enqueue bigint;
    v_scheduler bigint;
    v_consumer bigint;
    v_disposition_sha text;
    v_fence_sha text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Notifications V1 queue evidence principal denied';
    END IF;
    SELECT operation.cutover_generation INTO STRICT v_generation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
      AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
      AND operation.group_id='notifications' AND operation.domain='notifications';
    SELECT adoption.run_id,adoption.inventory_sha256,adoption.acceptance_sha256,
        adoption.item_count,adoption.expected_enqueue_fences,
        adoption.expected_scheduler_fences,adoption.expected_consumer_fences
    INTO STRICT v_async_run,v_inventory,v_acceptance,v_expected_items,
        v_expected_enqueue,v_expected_scheduler,v_expected_consumer
    FROM platform.async_v1_adoption_runs AS adoption
    WHERE adoption.scope_kind='tenant' AND adoption.scope_id=p_scope_id
      AND adoption.domain='notifications' AND adoption.group_id='notifications'
      AND adoption.target_generation=v_generation AND adoption.state='accepted';
    SELECT zero.receipt_sha256,zero.inventory_sha256
    INTO STRICT v_zero,v_zero_inventory
    FROM platform.async_v1_zero_inventory_receipts AS zero
    WHERE zero.run_id=v_async_run AND zero.queue_count=0 AND zero.timer_count=0
      AND zero.future_count=0 AND zero.running_count=0;
    SELECT pg_catalog.count(*) INTO v_items
    FROM platform.async_v1_adoption_items AS item WHERE item.run_id=v_async_run;
    SELECT pg_catalog.count(*),pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
            disposition.receipt_sha256,',' ORDER BY disposition.inventory_class,
            disposition.source_name,disposition.source_item_id),''),'UTF8'
        )),'hex')
    INTO v_dispositions,v_disposition_sha
    FROM platform.async_v1_adoption_dispositions AS disposition
    WHERE disposition.run_id=v_async_run;
    SELECT pg_catalog.count(*) FILTER (WHERE fence.fence_kind='enqueue'),
        pg_catalog.count(*) FILTER (WHERE fence.fence_kind='scheduler'),
        pg_catalog.count(*) FILTER (WHERE fence.fence_kind='consumer'),
        pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE(
            pg_catalog.string_agg(fence.receipt_sha256,',' ORDER BY fence.fence_kind,
                fence.component_name),''),'UTF8')),'hex')
    INTO v_enqueue,v_scheduler,v_consumer,v_fence_sha
    FROM platform.async_v1_restart_fence_receipts AS fence
    WHERE fence.run_id=v_async_run;
    IF v_zero_inventory IS DISTINCT FROM v_inventory OR
       v_items IS DISTINCT FROM v_expected_items OR
       v_dispositions IS DISTINCT FROM v_expected_items OR
       v_enqueue IS DISTINCT FROM v_expected_enqueue::bigint OR
       v_scheduler IS DISTINCT FROM v_expected_scheduler::bigint OR
       v_consumer IS DISTINCT FROM v_expected_consumer::bigint THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 queue adoption evidence differs';
    END IF;
    RETURN pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'harpy/notifications/v1-queue-evidence/v1:'||p_operation_id::text||':'||
        p_run_id::text||':'||v_async_run::text||':'||v_generation::text||':'||
        v_inventory||':'||v_acceptance||':'||v_zero||':'||v_disposition_sha||':'||
        v_fence_sha||':'||v_expected_items::text,'UTF8'
    )),'hex');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_queue_evidence_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
