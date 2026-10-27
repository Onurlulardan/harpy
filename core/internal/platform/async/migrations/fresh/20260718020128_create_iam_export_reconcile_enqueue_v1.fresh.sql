-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_export_reconcile_enqueue_v1(
    p_tenant_id uuid,p_export_id uuid,p_original_job_id uuid,
    p_resource_kind text,p_enqueued_at timestamptz
)
RETURNS TABLE(job_id uuid,replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_original platform.jobs%ROWTYPE;
    v_existing platform.jobs%ROWTYPE;
    v_job_id uuid;
    v_payload_ref text;
    v_payload_sha text;
    v_dedupe text;
    v_effect_key text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_original_job_id) IS DISTINCT FROM 7 OR
       p_resource_kind NOT IN ('users','roles','organizations') OR
       NOT pg_catalog.isfinite(p_enqueued_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export reconciliation enqueue authority is invalid';
    END IF;
    SELECT original.* INTO v_original FROM platform.jobs AS original
    WHERE original.id=p_original_job_id AND original.tenant_id=p_tenant_id
      AND original.scope_kind='tenant' AND original.scope_id=p_tenant_id
      AND original.authority_group_id='iam_system_resources'
      AND original.authority_domain='iam' AND original.writer_origin='v2'
      AND original.kind='iam.system-resource.record-export'
      AND original.payload_revision=1 AND original.effect_class='external'
      AND original.payload_ref='iam-system-record/'||p_resource_kind||'/'||
          p_export_id::text
      AND original.dedupe_key='iam-system-transfer:'||
          'iam.system-resource.record-export:'||p_export_id::text
      AND original.effect_idempotency_key='iam-system-export:'||p_export_id::text
      AND original.max_attempts=3 AND original.state='dead'
      AND original.last_error_class IS NULL
      AND original.last_error_code='JOB_OUTCOME_UNKNOWN'
      AND original.last_error_sha256 IS NULL
      AND original.completed_at=p_enqueued_at
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export unknown source is unavailable';
    END IF;
    v_payload_ref := 'iam-system-record/'||p_resource_kind||'/'||p_export_id::text;
    v_payload_sha := pg_catalog.encode(
        pg_catalog.sha256(pg_catalog.convert_to(v_payload_ref,'UTF8')),'hex');
    v_dedupe := 'iam-system-transfer:iam.system-resource.export-reconcile:'||
        p_export_id::text;
    v_effect_key := 'iam-system-export-reconcile:'||p_export_id::text;
    SELECT candidate.* INTO v_existing FROM platform.jobs AS candidate
    WHERE candidate.scope_kind='tenant' AND candidate.scope_id=p_tenant_id
      AND candidate.authority_group_id='iam_system_resources'
      AND candidate.authority_domain='iam'
      AND candidate.kind='iam.system-resource.export-reconcile'
      AND candidate.dedupe_key=v_dedupe FOR SHARE;
    IF FOUND THEN
        IF v_existing.tenant_id=p_tenant_id AND
           v_existing.authority_group_generation=v_original.authority_group_generation AND
           v_existing.cutover_generation=v_original.cutover_generation AND
           v_existing.writer_origin='v2' AND v_existing.payload_revision=1 AND
           v_existing.payload_ref=v_payload_ref AND
           v_existing.payload_sha256=v_payload_sha AND
           v_existing.effect_class='external' AND
           v_existing.effect_idempotency_key=v_effect_key AND
           v_existing.max_attempts=8 AND v_existing.created_at=p_enqueued_at THEN
            RETURN QUERY SELECT v_existing.id,true;
            RETURN;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation replay differs';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        v_original.scope_kind,v_original.scope_id,v_original.authority_group_id,
        v_original.authority_group_generation,v_original.authority_domain,
        v_original.cutover_generation
    );
    v_job_id := pg_catalog.uuidv7();
    INSERT INTO platform.jobs(
        id,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,
        writer_origin,kind,payload_revision,payload_ref,payload_sha256,dedupe_key,
        effect_class,effect_idempotency_key,priority,run_at,state,attempts,
        max_attempts,lease_generation,created_at,updated_at
    ) VALUES(
        v_job_id,p_tenant_id,'tenant',p_tenant_id,'iam_system_resources',
        v_original.authority_group_generation,'iam',v_original.cutover_generation,
        'v2','iam.system-resource.export-reconcile',1,v_payload_ref,v_payload_sha,
        v_dedupe,'external',v_effect_key,90,p_enqueued_at,'queued',0,8,0,
        p_enqueued_at,p_enqueued_at
    );
    RETURN QUERY SELECT v_job_id,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_iam_export_reconcile_enqueue_v1(
    uuid,uuid,uuid,text,timestamptz
) FROM PUBLIC;
