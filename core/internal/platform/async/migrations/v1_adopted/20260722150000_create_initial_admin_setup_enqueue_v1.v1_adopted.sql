-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_initial_admin_setup_enqueue_v1(
    p_operation_id uuid, p_tenant_id uuid, p_artifact_id uuid,
    p_artifact_ref text, p_schema_name name, p_manifest_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_digest text;
    v_effect_key text;
    v_now timestamptz := pg_catalog.statement_timestamp();
    v_inserted integer;
    v_existing platform.jobs%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_tenant_id IS NULL
       OR p_artifact_id IS NULL OR p_schema_name IS NULL
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$'
       OR p_artifact_ref IS DISTINCT FROM
          'iam-lifecycle-delivery://tenant/' || p_tenant_id::text ||
          '/artifact/' || p_artifact_id::text
       OR NOT platform.tenancy_activation_assert_v2(
           p_operation_id,p_tenant_id,p_schema_name
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='initial admin setup delivery enqueue denied';
    END IF;
    v_digest := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to('iam.lifecycle.delivery.v1','UTF8') ||
        '\x00'::bytea || pg_catalog.convert_to(p_artifact_ref,'UTF8')
    ), 'hex');
    v_effect_key := 'iam.lifecycle.delivery:' || p_artifact_ref;
    INSERT INTO platform.jobs (
        id,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,
        writer_origin,kind,payload_revision,payload_ref,payload_sha256,
        dedupe_key,effect_class,effect_idempotency_key,priority,run_at,
        state,attempts,max_attempts,lease_generation,retry_of_job_id
    ) VALUES (
        p_artifact_id,p_tenant_id,'tenant',p_tenant_id,
        'iam_system_resources',1,'iam',1,'v2','iam.lifecycle.delivery',1,
        p_artifact_ref,v_digest,p_artifact_id::text,'external',v_effect_key,
        100,v_now,'queued',0,8,0,NULL
    ) ON CONFLICT (id) DO NOTHING;
    GET DIAGNOSTICS v_inserted = ROW_COUNT;
    IF v_inserted=1 THEN
        RETURN false;
    END IF;
    SELECT * INTO v_existing FROM platform.jobs AS job
    WHERE job.id=p_artifact_id FOR SHARE;
    IF NOT FOUND OR ROW(v_existing.tenant_id,v_existing.scope_kind,
        v_existing.scope_id,v_existing.authority_group_id,
        v_existing.authority_group_generation,v_existing.authority_domain,
        v_existing.cutover_generation,v_existing.writer_origin,
        v_existing.kind,v_existing.payload_revision,v_existing.payload_ref,
        v_existing.payload_sha256,v_existing.dedupe_key,
        v_existing.effect_class,v_existing.effect_idempotency_key)
       IS DISTINCT FROM ROW(p_tenant_id,'tenant',p_tenant_id,
        'iam_system_resources',1::bigint,'iam',1::bigint,'v2',
        'iam.lifecycle.delivery',1,p_artifact_ref,v_digest,
        p_artifact_id::text,'external',v_effect_key) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='initial admin setup delivery replay conflicts';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v1(
    uuid,uuid,uuid,text,name,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v1(
    uuid,uuid,uuid,text,name,text
) TO harpy_runtime;
