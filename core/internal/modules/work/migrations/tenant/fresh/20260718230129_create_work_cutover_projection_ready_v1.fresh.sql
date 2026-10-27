-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_v1_adoption_projection_ready_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_scope_id uuid
)
RETURNS pg_lsn
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_expected bigint;
    v_complete bigint;
    v_final_count bigint;
    v_final_lsn pg_lsn;
    v_existing bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Work V1 projection readiness principal denied';
    END IF;
    PERFORM 1
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
      AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
      AND operation.group_id='work' AND operation.domain='work'
      AND operation.state='quiescing'
    FOR SHARE;
    IF NOT FOUND OR NOT EXISTS (
        SELECT 1 FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.tenant_id=p_scope_id
          AND identity.origin='v1_adopted'
    ) OR NOT EXISTS (
        SELECT 1 FROM work_adoption_runs AS run
        WHERE run.run_id=p_run_id AND run.state IN ('COPYING','RECONCILING')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'Work V1 projection authority is not ready';
    END IF;
    SELECT pg_catalog.count(*) INTO v_expected
    FROM platform.cutover_operation_checkpoint_keys AS expected
    WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id;
    SELECT pg_catalog.count(*), pg_catalog.count(DISTINCT checkpoint.final_lsn::text)
    INTO v_complete, v_final_count
    FROM platform.cutover_copy_checkpoints AS checkpoint
    WHERE checkpoint.operation_id=p_operation_id AND checkpoint.run_id=p_run_id
      AND checkpoint.state='complete' AND checkpoint.final_lsn IS NOT NULL;
    IF v_expected < 1 OR v_complete IS DISTINCT FROM v_expected OR
       v_final_count IS DISTINCT FROM 1::bigint OR EXISTS (
        SELECT 1
        FROM platform.cutover_operation_checkpoint_keys AS expected
        LEFT JOIN platform.cutover_copy_checkpoints AS checkpoint
          ON checkpoint.operation_id=expected.operation_id
         AND checkpoint.run_id=expected.run_id
         AND checkpoint.relation_ordinal=expected.relation_ordinal
         AND checkpoint.checkpoint_key=expected.checkpoint_key
         AND checkpoint.state='complete'
        WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id
          AND checkpoint.operation_id IS NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'Work V1 projection checkpoints are incomplete';
    END IF;
    SELECT checkpoint.final_lsn INTO STRICT v_final_lsn
    FROM platform.cutover_copy_checkpoints AS checkpoint
    WHERE checkpoint.operation_id=p_operation_id AND checkpoint.run_id=p_run_id
    ORDER BY checkpoint.relation_ordinal, checkpoint.checkpoint_key
    LIMIT 1;
    SELECT
        (SELECT pg_catalog.count(*) FROM work_boards) +
        (SELECT pg_catalog.count(*) FROM work_board_columns) +
        (SELECT pg_catalog.count(*) FROM work_board_memberships) +
        (SELECT pg_catalog.count(*) FROM work_tasks) +
        (SELECT pg_catalog.count(*) FROM work_task_assignees) +
        (SELECT pg_catalog.count(*) FROM work_labels) +
        (SELECT pg_catalog.count(*) FROM work_task_labels) +
        (SELECT pg_catalog.count(*) FROM work_checklist_items) +
        (SELECT pg_catalog.count(*) FROM work_comments) +
        (SELECT pg_catalog.count(*) FROM work_calendar_events) +
        (SELECT pg_catalog.count(*) FROM work_recurrence_rules) +
        (SELECT pg_catalog.count(*) FROM work_resource_links) +
        (SELECT pg_catalog.count(*) FROM work_inbox_items)
    INTO v_existing;
    IF v_existing <> 0 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'Work V1 inactive target is not empty';
    END IF;
    RETURN v_final_lsn;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION work_v1_adoption_projection_ready_v1(
    uuid, uuid, uuid
) FROM PUBLIC;
