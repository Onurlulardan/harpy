-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_accept_v1(
    p_operation_id uuid,p_cutover_run_id uuid,p_scope_id uuid
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_adoption_run_id uuid;
    v_final_lsn pg_lsn;
    v_binding_root text;
    v_effect_root text;
    v_row_root text;
    v_generation bigint;
    v_manifest text;
    v_source_count bigint;
    v_target_count bigint;
    v_source_bytes bigint;
    v_target_bytes bigint;
    v_source_root text;
    v_target_root text;
    v_acceptance bytea;
    v_receipt_sha text;
    v_rows bigint;
    v_now timestamptz:=pg_catalog.transaction_timestamp();
BEGIN
    SELECT ready.adoption_run_id,ready.final_lsn,
        ready.binding_root_sha256,ready.effect_root_sha256
    INTO STRICT v_adoption_run_id,v_final_lsn,v_binding_root,v_effect_root
    FROM documents_v1_adoption_projection_ready_v1(
        p_operation_id,p_cutover_run_id,p_scope_id
    ) AS ready;
    v_row_root:=documents_v1_adoption_project_rows_v1(
        p_operation_id,p_cutover_run_id,p_scope_id
    );
    SELECT operation.cutover_generation,operation.lease_manifest_sha256
    INTO STRICT v_generation,v_manifest
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id
      AND operation.run_id=p_cutover_run_id
    FOR SHARE;
    SELECT pg_catalog.count(*)::bigint,
        pg_catalog.count(*) FILTER (
            WHERE checkpoint.lifecycle_state='VERIFIED'
        )::bigint,
        COALESCE(pg_catalog.sum(checkpoint.source_bytes),0)::bigint,
        COALESCE(pg_catalog.sum(checkpoint.source_bytes) FILTER (
            WHERE checkpoint.lifecycle_state='VERIFIED'
        ),0)::bigint,
        pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE(
            pg_catalog.string_agg(checkpoint.source_version_id::text||':'||
                checkpoint.source_object_revision||':'||
                pg_catalog.encode(checkpoint.source_sha256,'hex')||':'||
                pg_catalog.encode(checkpoint.source_metadata_sha256,'hex'),','
                ORDER BY checkpoint.source_version_id,
                         checkpoint.source_object_revision),''),'UTF8')),'hex'),
        pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE(
            pg_catalog.string_agg(checkpoint.source_version_id::text||':'||
                checkpoint.source_object_revision||':'||
                pg_catalog.encode(checkpoint.target_plaintext_sha256,'hex')||':'||
                pg_catalog.encode(checkpoint.metadata_sha256,'hex'),','
                ORDER BY checkpoint.target_version_id,
                         checkpoint.source_object_revision),''),'UTF8')),'hex')
    INTO v_source_count,v_target_count,v_source_bytes,v_target_bytes,
         v_source_root,v_target_root
    FROM documents_blob_copy_checkpoints AS checkpoint
    WHERE checkpoint.run_id=v_adoption_run_id;
    IF v_source_count IS DISTINCT FROM v_target_count OR
       v_source_bytes IS DISTINCT FROM v_target_bytes OR
       v_source_root IS DISTINCT FROM v_target_root OR
       v_target_count IS DISTINCT FROM (
           SELECT pg_catalog.count(*)
           FROM documents_v1_adoption_stage AS stage
           WHERE stage.operation_id=p_operation_id
             AND stage.cutover_run_id=p_cutover_run_id
             AND stage.target_kind='version' AND NOT stage.removed
       ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 exact digest reconciliation differs';
    END IF;
    v_acceptance:=pg_catalog.sha256(
        pg_catalog.decode(v_manifest,'hex')||
        pg_catalog.decode(v_source_root,'hex')||
        pg_catalog.decode(v_target_root,'hex')||
        pg_catalog.decode(v_row_root,'hex')||
        pg_catalog.decode(v_binding_root,'hex')||
        pg_catalog.decode(v_effect_root,'hex')
    );
    UPDATE documents_adoption_runs SET lifecycle_state='ACCEPTED',
        source_count=v_source_count,target_count=v_target_count,
        source_plaintext_bytes=v_source_bytes,
        target_plaintext_bytes=v_target_bytes,
        source_merkle_sha256=pg_catalog.decode(v_source_root,'hex'),
        target_merkle_sha256=pg_catalog.decode(v_target_root,'hex'),
        pending_object_count=0,pending_job_count=0,
        final_lsn=v_final_lsn,applied_lsn=v_final_lsn,
        acceptance_sha256=v_acceptance,updated_at=v_now
    WHERE run_id=v_adoption_run_id AND cutover_generation=v_generation
      AND manifest_sha256=pg_catalog.decode(v_manifest,'hex')
      AND lifecycle_state IN (
          'INVENTORY','COPYING','SHADOW','RECONCILING','QUIESCED'
      );
    GET DIAGNOSTICS v_rows=ROW_COUNT;
    IF v_rows IS DISTINCT FROM 1::bigint THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 adoption acceptance state differs';
    END IF;
    v_receipt_sha:=pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'harpy/documents/v1-cutover-receipt/v1:'||p_operation_id::text||':'||
        p_cutover_run_id::text||':'||v_adoption_run_id::text||':'||p_scope_id::text||
        ':'||v_generation::text||':'||v_source_count::text||':'||
        v_source_bytes::text||':'||v_source_root||':'||v_target_root||':'||
        v_row_root||':'||v_binding_root||':'||v_effect_root||':'||v_final_lsn::text,
        'UTF8'
    )),'hex');
    INSERT INTO documents_v1_cutover_receipts (
        operation_id,cutover_run_id,adoption_run_id,tenant_id,cutover_generation,
        source_count,target_count,source_plaintext_bytes,target_plaintext_bytes,
        source_merkle_sha256,target_merkle_sha256,row_root_sha256,
        binding_root_sha256,effect_root_sha256,final_lsn,accepted_at,receipt_sha256
    ) VALUES (
        p_operation_id,p_cutover_run_id,v_adoption_run_id,p_scope_id,v_generation,
        v_source_count,v_target_count,v_source_bytes,v_target_bytes,
        v_source_root,v_target_root,v_row_root,v_binding_root,v_effect_root,
        v_final_lsn,v_now,v_receipt_sha
    );
    RETURN v_row_root;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_accept_v1(uuid,uuid,uuid)
FROM PUBLIC;
