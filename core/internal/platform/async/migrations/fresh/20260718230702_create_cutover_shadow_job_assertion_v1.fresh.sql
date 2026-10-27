-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_assert_job_row_authority_v1(
    p_job platform.jobs
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_bound boolean;
BEGIN
    SELECT pg_catalog.count(*)=1 INTO v_bound
    FROM platform.cutover_job_bindings AS binding
    WHERE binding.job_id=p_job.id;
    IF v_bound THEN
        PERFORM platform.platform_async_assert_cutover_shadow_authority_v1(
            p_job.scope_kind,p_job.scope_id,p_job.authority_group_id,
            p_job.authority_group_generation,p_job.authority_domain,
            p_job.cutover_generation);
        IF p_job.scope_kind IS DISTINCT FROM 'tenant' OR
           p_job.tenant_id IS DISTINCT FROM p_job.scope_id OR
           p_job.writer_origin IS DISTINCT FROM 'v2' OR
           p_job.kind IS DISTINCT FROM 'documents.adoption.copy' OR
           p_job.payload_revision IS DISTINCT FROM 1 OR
           p_job.effect_class IS DISTINCT FROM 'external' OR
           p_job.payload_ref IS NULL OR p_job.effect_idempotency_key IS NULL OR
           NOT EXISTS (
               SELECT 1 FROM platform.cutover_job_bindings AS binding
               WHERE binding.job_id=p_job.id
                 AND binding.tenant_id=p_job.scope_id
                 AND binding.group_id=p_job.authority_group_id
                 AND binding.group_generation=p_job.authority_group_generation
                 AND binding.domain=p_job.authority_domain
                 AND binding.cutover_generation=p_job.cutover_generation
                 AND binding.job_kind=p_job.kind
                 AND binding.payload_revision=p_job.payload_revision
                 AND binding.payload_ref=p_job.payload_ref
                 AND binding.payload_sha256=p_job.payload_sha256
                 AND binding.effect_idempotency_key=p_job.effect_idempotency_key
           ) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5603',
                MESSAGE = 'cutover shadow job binding differs';
        END IF;
        RETURN;
    END IF;
    IF p_job.kind='documents.adoption.copy' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'Documents adoption copy job is unbound';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        p_job.scope_kind,p_job.scope_id,p_job.authority_group_id,
        p_job.authority_group_generation,p_job.authority_domain,
        p_job.cutover_generation);
END;
$function$;
CREATE FUNCTION platform.platform_async_assert_job_id_authority_v1(p_job_id uuid)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
BEGIN
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603', MESSAGE = 'job is missing';
    END IF;
    PERFORM platform.platform_async_assert_job_row_authority_v1(v_job);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_assert_job_row_authority_v1(
    platform.jobs) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_assert_job_id_authority_v1(uuid)
FROM PUBLIC;
