-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_v1_adoption_project_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_scope_id uuid
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_kind text;
    v_stage work_v1_adoption_stage%ROWTYPE;
    v_final_lsn pg_lsn;
    v_source_count bigint;
    v_source_identity_count bigint;
    v_target_count bigint;
    v_source_sha text;
    v_target_sha text;
    v_manifest bytea;
    v_acceptance bytea;
    v_actor uuid;
    v_generation bigint;
    v_rows bigint;
BEGIN
    v_target_count := 0;
    v_final_lsn := work_v1_adoption_projection_ready_v1(
        p_operation_id, p_run_id, p_scope_id
    );
    SELECT operation.lease_owner, operation.cutover_generation
    INTO STRICT v_actor, v_generation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id;
    FOREACH v_kind IN ARRAY ARRAY[
        'BOARD','COLUMN','MEMBERSHIP','TASK','LABEL','ASSIGNEE',
        'TASK_LABEL','CHECKLIST','COMMENT','CALENDAR','RECURRENCE',
        'FORM_LINK','WORKFLOW_PROJECTION'
    ]::text[]
    LOOP
        FOR v_stage IN
            SELECT stage.*
            FROM work_v1_adoption_stage AS stage
            WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
              AND stage.target_kind=v_kind AND NOT stage.removed
            ORDER BY stage.target_id
        LOOP
            IF v_kind IN ('BOARD','COLUMN','MEMBERSHIP') THEN
                PERFORM work_v1_adoption_project_board_v1(
                    v_stage.target_kind, v_stage.target_id, v_stage.target_record
                );
            ELSIF v_kind IN (
                'TASK','LABEL','ASSIGNEE','TASK_LABEL','CHECKLIST','FORM_LINK'
            ) THEN
                PERFORM work_v1_adoption_project_task_v1(
                    v_stage.target_kind, v_stage.target_id, v_stage.target_record
                );
            ELSE
                PERFORM work_v1_adoption_project_collaboration_v1(
                    v_stage.target_kind, v_stage.target_id, v_stage.target_record
                );
            END IF;
            v_target_count := v_target_count + 1;
        END LOOP;
    END LOOP;
    SELECT pg_catalog.count(*),
        pg_catalog.count(DISTINCT (source.source_kind, source.source_id)),
        pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
            source.source_kind || ':' || source.source_id::text || ':' ||
            pg_catalog.encode(source.source_sha256, 'hex'),
            ',' ORDER BY source.source_kind, source.source_id
        ), ''), 'UTF8')
    ), 'hex')
    INTO v_source_count, v_source_identity_count, v_source_sha
    FROM (
        SELECT DISTINCT stage.source_kind, stage.source_id, stage.source_sha256
        FROM work_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
    ) AS source;
    IF v_source_count IS DISTINCT FROM v_source_identity_count THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'Work V1 source identity digest differs';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        COALESCE(pg_catalog.string_agg(
            stage.target_kind || ':' || stage.target_id::text || ':' ||
            pg_catalog.encode(stage.target_sha256, 'hex'),
            ',' ORDER BY stage.target_kind, stage.target_id
        ), ''), 'UTF8'
    )), 'hex')
    INTO v_target_sha
    FROM work_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed;
    INSERT INTO work_adoption_objects (
        run_id, source_kind, source_id, target_kind, target_id,
        source_sha256, target_sha256, applied_lsn, state, issue_code, updated_at
    )
    SELECT p_run_id, stage.source_kind, stage.source_id, stage.target_kind,
        stage.target_id, stage.source_sha256, stage.target_sha256,
        v_final_lsn, 'MATCHED', NULL, transaction_timestamp()
    FROM work_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed
    ORDER BY stage.source_kind, stage.source_id, stage.target_kind, stage.target_id
    ON CONFLICT (run_id, source_kind, source_id, target_kind, target_id) DO UPDATE
    SET source_sha256=EXCLUDED.source_sha256,
        target_sha256=EXCLUDED.target_sha256,
        applied_lsn=EXCLUDED.applied_lsn, state='MATCHED',
        issue_code=NULL, updated_at=EXCLUDED.updated_at
    WHERE work_adoption_objects.source_sha256=EXCLUDED.source_sha256
      AND work_adoption_objects.target_sha256=EXCLUDED.target_sha256;
    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows IS DISTINCT FROM v_target_count THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'Work V1 adoption object reconciliation differs';
    END IF;
    SELECT run.manifest_sha256 INTO STRICT v_manifest
    FROM work_adoption_runs AS run
    WHERE run.run_id=p_run_id AND run.cutover_generation=v_generation
    FOR UPDATE;
    v_acceptance := pg_catalog.sha256(
        v_manifest || pg_catalog.decode(v_source_sha, 'hex') ||
        pg_catalog.decode(v_target_sha, 'hex')
    );
    UPDATE work_adoption_runs
    SET state='ACCEPTED', source_count=v_source_count,
        target_count=v_target_count,
        source_sha256=pg_catalog.decode(v_source_sha, 'hex'),
        target_sha256=pg_catalog.decode(v_target_sha, 'hex'),
        final_lsn=v_final_lsn, applied_lsn=v_final_lsn,
        acceptance_sha256=v_acceptance, updated_at=transaction_timestamp()
    WHERE run_id=p_run_id AND cutover_generation=v_generation
      AND state IN ('COPYING','RECONCILING');
    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows IS DISTINCT FROM 1::bigint THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'Work V1 adoption acceptance state differs';
    END IF;
    UPDATE work_v1_writer_fence
    SET state='V2_FENCED', run_id=p_run_id, generation=v_generation,
        reverse_reconciliation_id=NULL, updated_by=v_actor,
        updated_at=transaction_timestamp()
    WHERE singleton AND state='V1_ACTIVE';
    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows IS DISTINCT FROM 1::bigint THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'Work V1 writer fence differs';
    END IF;
    RETURN v_target_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION work_v1_adoption_project_v1(
    uuid, uuid, uuid
) FROM PUBLIC;
