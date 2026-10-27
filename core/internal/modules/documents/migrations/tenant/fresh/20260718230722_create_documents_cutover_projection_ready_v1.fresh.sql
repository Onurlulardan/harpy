-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_projection_ready_v1(
    p_operation_id uuid,p_cutover_run_id uuid,p_scope_id uuid
)
RETURNS TABLE (
    adoption_run_id uuid,final_lsn pg_lsn,
    binding_root_sha256 text,effect_root_sha256 text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_expected bigint;
    v_complete bigint;
    v_final_count bigint;
    v_checkpoint_count bigint;
    v_adoption_run_id uuid;
    v_final_lsn pg_lsn;
    v_binding_root text;
    v_effect_root text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Documents V1 projection principal denied';
    END IF;
    PERFORM 1 FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id
      AND operation.run_id=p_cutover_run_id
      AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
      AND operation.group_id='documents' AND operation.domain='documents'
      AND operation.state='quiescing'
      AND operation.lease_expires_at>pg_catalog.clock_timestamp()
    FOR SHARE;
    IF NOT FOUND OR NOT EXISTS (
        SELECT 1 FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.tenant_id=p_scope_id
          AND identity.schema_name=pg_catalog.current_schema()
          AND identity.origin='v1_adopted'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 projection authority is not ready';
    END IF;
    v_adoption_run_id:=documents_v1_adoption_mapping_ready_v1(
        p_operation_id,p_cutover_run_id
    );
    SELECT pg_catalog.count(*) INTO v_expected
    FROM platform.cutover_operation_checkpoint_keys AS expected
    WHERE expected.operation_id=p_operation_id
      AND expected.run_id=p_cutover_run_id;
    SELECT pg_catalog.count(*),
        pg_catalog.count(DISTINCT checkpoint.final_lsn::text)
    INTO v_complete,v_final_count
    FROM platform.cutover_copy_checkpoints AS checkpoint
    WHERE checkpoint.operation_id=p_operation_id
      AND checkpoint.run_id=p_cutover_run_id
      AND checkpoint.state='complete' AND checkpoint.final_lsn IS NOT NULL;
    IF v_expected<1 OR v_complete IS DISTINCT FROM v_expected OR
       v_final_count IS DISTINCT FROM 1::bigint OR EXISTS (
        SELECT 1
        FROM platform.cutover_operation_checkpoint_keys AS expected
        LEFT JOIN platform.cutover_copy_checkpoints AS checkpoint
          ON checkpoint.operation_id=expected.operation_id
         AND checkpoint.run_id=expected.run_id
         AND checkpoint.relation_ordinal=expected.relation_ordinal
         AND checkpoint.checkpoint_key=expected.checkpoint_key
         AND checkpoint.state='complete'
        WHERE expected.operation_id=p_operation_id
          AND expected.run_id=p_cutover_run_id
          AND checkpoint.operation_id IS NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 logical checkpoints are incomplete';
    END IF;
    SELECT checkpoint.final_lsn INTO STRICT v_final_lsn
    FROM platform.cutover_copy_checkpoints AS checkpoint
    WHERE checkpoint.operation_id=p_operation_id
      AND checkpoint.run_id=p_cutover_run_id
    ORDER BY checkpoint.relation_ordinal,checkpoint.checkpoint_key LIMIT 1;
    SELECT pg_catalog.count(*) INTO v_checkpoint_count
    FROM documents_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id
      AND stage.cutover_run_id=p_cutover_run_id
      AND stage.target_kind='checkpoint' AND NOT stage.removed;
    IF (SELECT pg_catalog.count(*) FROM documents_blob_copy_checkpoints
        WHERE run_id=v_adoption_run_id) IS DISTINCT FROM v_checkpoint_count OR
       EXISTS (
        SELECT 1 FROM documents_v1_adoption_stage AS stage
        LEFT JOIN documents_blob_copy_checkpoints AS checkpoint
          ON checkpoint.run_id=v_adoption_run_id
         AND checkpoint.source_version_id=
             (stage.target_record->>'source_version_id')::uuid
         AND checkpoint.source_object_revision=
             stage.target_record->>'source_object_revision'
         AND checkpoint.target_version_id=stage.target_id
         AND checkpoint.source_sha256=pg_catalog.decode(
             stage.target_record->>'source_sha256','hex')
         AND checkpoint.source_metadata_sha256=pg_catalog.decode(
             stage.target_record->>'source_metadata_sha256','hex')
         AND checkpoint.source_lsn=(stage.target_record->>'source_lsn')::pg_lsn
         AND checkpoint.lifecycle_state='VERIFIED'
         AND checkpoint.applied_lsn=checkpoint.source_lsn
         AND checkpoint.target_plaintext_sha256=checkpoint.source_sha256
         AND checkpoint.metadata_sha256=checkpoint.source_metadata_sha256
        WHERE stage.operation_id=p_operation_id
          AND stage.cutover_run_id=p_cutover_run_id
          AND stage.target_kind='checkpoint' AND NOT stage.removed
          AND checkpoint.run_id IS NULL
    ) OR EXISTS (
        SELECT 1 FROM documents_blob_copy_checkpoints AS checkpoint
        WHERE checkpoint.run_id=v_adoption_run_id
          AND (checkpoint.lifecycle_state<>'VERIFIED' OR
               checkpoint.applied_lsn IS NULL OR
               checkpoint.applied_lsn>v_final_lsn)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 object checkpoints are incomplete';
    END IF;
    IF (SELECT pg_catalog.count(*) FROM platform.cutover_job_bindings AS binding
        WHERE binding.operation_id=p_operation_id
          AND binding.cutover_run_id=p_cutover_run_id)
       IS DISTINCT FROM v_checkpoint_count OR EXISTS (
        SELECT 1 FROM platform.cutover_job_bindings AS binding
        JOIN platform.jobs AS job ON job.id=binding.job_id
        WHERE binding.operation_id=p_operation_id
          AND binding.cutover_run_id=p_cutover_run_id
          AND (binding.tenant_id<>p_scope_id OR
               binding.adoption_run_id<>v_adoption_run_id OR
               binding.job_kind<>'documents.adoption.copy' OR
               job.state<>'succeeded' OR job.kind<>binding.job_kind OR
               NOT EXISTS (
                   SELECT 1 FROM platform.job_attempts AS attempt
                   JOIN platform.idempotency_receipts AS receipt
                     ON receipt.receipt_id=attempt.effect_receipt_id
                   WHERE attempt.job_id=job.id AND attempt.outcome='succeeded'
                     AND attempt.finished_at IS NOT NULL
                     AND receipt.causation_id=job.id
                     AND receipt.effect_kind='documents.adoption.copy'
                     AND receipt.provider='documents.adoption'
                     AND receipt.operation='copy'
                     AND receipt.effect_status='succeeded'
               ))
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 shadow jobs or effects are pending';
    END IF;
    IF (SELECT (SELECT pg_catalog.count(*) FROM documents_folders)+
               (SELECT pg_catalog.count(*) FROM documents_items)+
               (SELECT pg_catalog.count(*) FROM documents_versions)+
               (SELECT pg_catalog.count(*) FROM documents_v1_cutover_receipts))<>0 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 inactive target is not empty';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE(
        pg_catalog.string_agg(binding.job_id::text||':'||binding.source_version_id::text||
            ':'||binding.payload_sha256,',' ORDER BY binding.job_id),''),'UTF8')),'hex')
    INTO v_binding_root FROM platform.cutover_job_bindings AS binding
    WHERE binding.operation_id=p_operation_id
      AND binding.cutover_run_id=p_cutover_run_id;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE(
        pg_catalog.string_agg(receipt.causation_id::text||':'||receipt.effect_id::text||
            ':'||receipt.effect_result_sha256,',' ORDER BY receipt.causation_id),''),
        'UTF8')),'hex') INTO v_effect_root
    FROM platform.idempotency_receipts AS receipt
    JOIN platform.cutover_job_bindings AS binding
      ON binding.job_id=receipt.causation_id
    WHERE binding.operation_id=p_operation_id
      AND binding.cutover_run_id=p_cutover_run_id
      AND receipt.effect_status='succeeded';
    RETURN QUERY SELECT v_adoption_run_id,v_final_lsn,v_binding_root,v_effect_root;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_projection_ready_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
