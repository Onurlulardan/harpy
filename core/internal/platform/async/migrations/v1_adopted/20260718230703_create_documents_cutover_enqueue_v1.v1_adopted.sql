-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_enqueue_documents_cutover_copy_v1(
    p_operation_id uuid, p_cutover_run_id uuid, p_adoption_run_id uuid,
    p_source_version_id uuid, p_job_id uuid, p_enqueued_at timestamptz
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_group_generation bigint;
    v_payload_ref text;
    v_payload_sha256 text;
    v_effect_key text;
    v_document bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       pg_catalog.uuid_extract_version(p_adoption_run_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_source_version_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       p_enqueued_at IS NULL OR p_enqueued_at>pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Documents cutover enqueue authority is invalid';
    END IF;
    SELECT operation.* INTO v_operation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id
      AND operation.run_id=p_cutover_run_id
      AND operation.scope_kind='tenant'
      AND operation.group_id='documents' AND operation.domain='documents'
      AND operation.state IN ('copying','shadow')
      AND operation.lease_expires_at>pg_catalog.clock_timestamp()
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5601',
            MESSAGE = 'Documents cutover operation is stale';
    END IF;
    SELECT group_row.generation INTO v_group_generation
    FROM platform.cutover_groups AS group_row
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=group_row.scope_kind
     AND domain_row.scope_id=group_row.scope_id
     AND domain_row.domain='documents'
    JOIN platform.cutover_group_members AS member
      ON member.scope_kind=group_row.scope_kind
     AND member.scope_id=group_row.scope_id
     AND member.group_id=group_row.group_id
     AND member.domain=domain_row.domain
    WHERE group_row.scope_kind='tenant'
      AND group_row.scope_id=v_operation.scope_id
      AND group_row.group_id='documents'
      AND group_row.state IN ('v1','shadow') AND group_row.writer='v1'
      AND domain_row.state IN ('v1','shadow') AND domain_row.writer='v1'
      AND domain_row.generation=v_operation.cutover_generation
    FOR SHARE OF group_row,domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5602',
            MESSAGE = 'Documents shadow writer authority is stale';
    END IF;
    v_payload_ref := 'documents/v1/adoption-copy/'||v_operation.scope_id::text||
        '/'||p_adoption_run_id::text||'/'||p_source_version_id::text||'/'||
        v_operation.cutover_generation::text||'/'||v_operation.lease_manifest_sha256;
    v_document := pg_catalog.convert_to('{"run_id":"'||p_adoption_run_id::text||
        '","source_version_id":"'||p_source_version_id::text||
        '","cutover_generation":'||v_operation.cutover_generation::text||
        ',"manifest_sha256":"'||v_operation.lease_manifest_sha256||'"}','UTF8');
    v_payload_sha256 := pg_catalog.encode(pg_catalog.sha256(v_document),'hex');
    v_effect_key := 'documents:adoption-copy:'||p_adoption_run_id::text||':'||
        p_source_version_id::text;
    INSERT INTO platform.cutover_job_bindings (
        job_id,operation_id,cutover_run_id,tenant_id,group_id,group_generation,
        domain,cutover_generation,lease_manifest_sha256,adoption_run_id,
        source_version_id,job_kind,payload_revision,payload_ref,payload_sha256,
        effect_idempotency_key,enqueued_at
    ) VALUES (
        p_job_id,p_operation_id,p_cutover_run_id,v_operation.scope_id,
        'documents',v_group_generation,'documents',v_operation.cutover_generation,
        v_operation.lease_manifest_sha256,p_adoption_run_id,p_source_version_id,
        'documents.adoption.copy',1,v_payload_ref,v_payload_sha256,v_effect_key,
        p_enqueued_at
    ) ON CONFLICT DO NOTHING;
    INSERT INTO platform.jobs (
        id,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,
        writer_origin,kind,payload_revision,payload_ref,payload_sha256,dedupe_key,
        effect_class,effect_idempotency_key,priority,run_at,state,attempts,
        max_attempts,lease_generation
    ) VALUES (
        p_job_id,v_operation.scope_id,'tenant',v_operation.scope_id,'documents',
        v_group_generation,'documents',v_operation.cutover_generation,'v2',
        'documents.adoption.copy',1,v_payload_ref,v_payload_sha256,v_effect_key,
        'external',v_effect_key,100,p_enqueued_at,'queued',0,8,0
    ) ON CONFLICT DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM platform.cutover_job_bindings AS binding
        JOIN platform.jobs AS job ON job.id=binding.job_id
        WHERE binding.job_id=p_job_id AND binding.operation_id=p_operation_id
          AND binding.cutover_run_id=p_cutover_run_id
          AND binding.adoption_run_id=p_adoption_run_id
          AND binding.source_version_id=p_source_version_id
          AND binding.payload_ref=v_payload_ref
          AND binding.payload_sha256=v_payload_sha256
          AND binding.effect_idempotency_key=v_effect_key
          AND job.scope_id=v_operation.scope_id AND job.kind=binding.job_kind
          AND job.payload_ref=binding.payload_ref
          AND job.payload_sha256=binding.payload_sha256
          AND job.effect_idempotency_key=binding.effect_idempotency_key
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'Documents cutover enqueue replay differs';
    END IF;
    RETURN p_job_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_enqueue_documents_cutover_copy_v1(
    uuid,uuid,uuid,uuid,uuid,timestamptz) FROM PUBLIC;
