-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_export_current_lease_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_lease_owner_id uuid,p_lease_generation bigint,p_resource_kind text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_lease_owner_id) IS DISTINCT FROM 7 OR
       p_lease_generation NOT BETWEEN 1 AND 3 OR
       p_resource_kind NOT IN ('users','roles','organizations') THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export lease authority is invalid';
    END IF;
    SELECT job.* INTO v_job
    FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_domain='iam' AND job.writer_origin='v2'
      AND job.authority_group_id='iam_system_resources'
      AND job.kind='iam.system-resource.record-export'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.effect_idempotency_key='iam-system-export:'||p_export_id::text
      AND job.dedupe_key='iam-system-transfer:'||
          'iam.system-resource.record-export:'||p_export_id::text
      AND job.payload_ref='iam-system-record/'||p_resource_kind||'/'||
          p_export_id::text
      AND job.max_attempts=3
    FOR SHARE;
    IF NOT FOUND OR v_job.state IS DISTINCT FROM 'running' OR
       v_job.attempts IS DISTINCT FROM p_lease_generation::integer OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_generation IS DISTINCT FROM p_lease_generation OR
       v_job.lease_until<=pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export lease is stale';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_group_generation,v_job.authority_domain,
        v_job.cutover_generation
    );
    SELECT attempt.* INTO v_attempt
    FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_lease_generation
      AND attempt.lease_generation=p_lease_generation
      AND attempt.lease_owner_id=p_lease_owner_id
    FOR SHARE;
    IF NOT FOUND OR v_attempt.finished_at IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export attempt is stale';
    END IF;
    RETURN v_job.cancel_requested_at IS NOT NULL;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_iam_export_current_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text
) FROM PUBLIC;
