-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_project_checkpoint_v1(
    p_operation_id uuid,p_cutover_run_id uuid,p_scope_id uuid,p_record jsonb
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_adoption_run_id uuid;
    v_source_version_id uuid;
    v_target_version_id uuid;
    v_job_id uuid;
    v_enqueued_at timestamptz;
    v_source_sha bytea;
    v_metadata_sha bytea;
    v_persisted_job uuid;
BEGIN
    v_adoption_run_id:=(p_record->>'adoption_run_id')::uuid;
    v_source_version_id:=(p_record->>'source_version_id')::uuid;
    v_target_version_id:=(p_record->>'target_version_id')::uuid;
    v_job_id:=(p_record->>'job_id')::uuid;
    v_enqueued_at:=(p_record->>'enqueued_at')::timestamptz;
    v_source_sha:=pg_catalog.decode(p_record->>'source_sha256','hex');
    v_metadata_sha:=pg_catalog.decode(p_record->>'source_metadata_sha256','hex');
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Documents V1 checkpoint projection principal denied';
    END IF;
    PERFORM 1 FROM documents_adoption_runs AS run
    JOIN platform.cutover_operations AS operation
      ON operation.operation_id=p_operation_id
     AND operation.run_id=p_cutover_run_id
     AND operation.cutover_generation=run.cutover_generation
     AND operation.lease_manifest_sha256=pg_catalog.encode(run.manifest_sha256,'hex')
    WHERE run.run_id=v_adoption_run_id
      AND run.lifecycle_state IN ('INVENTORY','COPYING','SHADOW')
      AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
      AND operation.group_id='documents' AND operation.domain='documents'
      AND operation.state IN ('copying','shadow')
      AND operation.lease_expires_at>pg_catalog.clock_timestamp()
    FOR SHARE OF run,operation;
    IF NOT FOUND OR NOT EXISTS (
        SELECT 1 FROM documents_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id
          AND stage.cutover_run_id=p_cutover_run_id
          AND stage.target_kind='version' AND stage.target_id=v_target_version_id
          AND NOT stage.removed
          AND (stage.target_record->>'item_id')::uuid=
              (p_record->>'target_document_id')::uuid
          AND stage.source_kind='version'
          AND stage.source_identity=(SELECT source.source_identity
              FROM documents_v1_adoption_stage AS source
              WHERE source.operation_id=p_operation_id
                AND source.cutover_run_id=p_cutover_run_id
                AND source.target_kind='checkpoint'
                AND source.target_id=v_target_version_id AND NOT source.removed)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 checkpoint mapping differs';
    END IF;
    INSERT INTO documents_blob_copy_checkpoints (
        run_id,source_document_id,source_version_id,source_object_key,
        source_object_revision,source_bytes,source_sha256,source_metadata_sha256,
        source_lsn,target_document_id,target_version_id,lifecycle_state,updated_at
    ) VALUES (
        v_adoption_run_id,(p_record->>'source_document_id')::uuid,
        v_source_version_id,p_record->>'source_object_key',
        p_record->>'source_object_revision',(p_record->>'source_bytes')::bigint,
        v_source_sha,v_metadata_sha,(p_record->>'source_lsn')::pg_lsn,
        (p_record->>'target_document_id')::uuid,v_target_version_id,
        'INVENTORIED',v_enqueued_at
    ) ON CONFLICT (run_id,source_version_id,source_object_revision) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM documents_blob_copy_checkpoints AS checkpoint
        WHERE checkpoint.run_id=v_adoption_run_id
          AND checkpoint.source_document_id=(p_record->>'source_document_id')::uuid
          AND checkpoint.source_version_id=v_source_version_id
          AND checkpoint.source_object_key=p_record->>'source_object_key'
          AND checkpoint.source_object_revision=p_record->>'source_object_revision'
          AND checkpoint.source_bytes=(p_record->>'source_bytes')::bigint
          AND checkpoint.source_sha256=v_source_sha
          AND checkpoint.source_metadata_sha256=v_metadata_sha
          AND checkpoint.source_lsn=(p_record->>'source_lsn')::pg_lsn
          AND checkpoint.target_document_id=(p_record->>'target_document_id')::uuid
          AND checkpoint.target_version_id=v_target_version_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 checkpoint replay differs';
    END IF;
    v_persisted_job:=platform.platform_async_enqueue_documents_cutover_copy_v1(
        p_operation_id,p_cutover_run_id,v_adoption_run_id,v_source_version_id,
        v_job_id,v_enqueued_at
    );
    IF v_persisted_job IS DISTINCT FROM v_job_id THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 checkpoint job identity differs';
    END IF;
    UPDATE documents_adoption_runs AS run SET
        source_count=stats.source_count,
        source_plaintext_bytes=stats.source_bytes,
        pending_object_count=stats.pending_objects,
        pending_job_count=stats.pending_jobs,
        updated_at=pg_catalog.transaction_timestamp()
    FROM (
        SELECT pg_catalog.count(*)::bigint AS source_count,
            COALESCE(pg_catalog.sum(checkpoint.source_bytes),0)::bigint AS source_bytes,
            pg_catalog.count(*) FILTER (WHERE checkpoint.lifecycle_state<>'VERIFIED')
                ::bigint AS pending_objects,
            pg_catalog.count(*) FILTER (WHERE job.state<>'succeeded')
                ::bigint AS pending_jobs
        FROM documents_blob_copy_checkpoints AS checkpoint
        JOIN platform.cutover_job_bindings AS binding
          ON binding.adoption_run_id=checkpoint.run_id
         AND binding.source_version_id=checkpoint.source_version_id
        JOIN platform.jobs AS job ON job.id=binding.job_id
        WHERE checkpoint.run_id=v_adoption_run_id
    ) AS stats WHERE run.run_id=v_adoption_run_id;
    RETURN v_job_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_project_checkpoint_v1(
    uuid,uuid,uuid,jsonb
) FROM PUBLIC;
