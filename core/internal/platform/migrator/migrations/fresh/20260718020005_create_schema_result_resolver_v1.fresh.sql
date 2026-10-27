-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_resolve_schema_result_v1(
    p_tenant_id uuid,
    p_job_id uuid,
    p_plan_sha256 text,
    p_minimum_generation bigint
)
RETURNS TABLE(
    job_id uuid,
    result_generation bigint,
    succeeded boolean,
    receipt_sha256 bytea,
    failure_code text,
    failure_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_job platform.schema_jobs%ROWTYPE;
    v_failure_digest text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.uuid_extract_version(p_tenant_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$'
       OR p_minimum_generation IS NULL
       OR p_minimum_generation NOT BETWEEN 1 AND 9007199254740991 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema result identity is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM platform.tenants AS tenant
        WHERE tenant.id = p_tenant_id
          AND tenant.deleted_at IS NULL
          AND tenant.status IN ('ACTIVE', 'TRIAL')
        FOR SHARE
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'schema result TenantScope is invalid';
    END IF;
    PERFORM platform.billing_resolve_current_tenant_authority_v1(p_tenant_id);
    SELECT candidate.* INTO v_job
    FROM platform.schema_jobs AS candidate
    WHERE candidate.tenant_id = p_tenant_id
      AND candidate.job_id = p_job_id
      AND candidate.owner = 'forms'
      AND candidate.plan_sha256 = p_plan_sha256
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'schema result binding differs';
    END IF;
    IF v_job.generation < p_minimum_generation
       OR v_job.state NOT IN ('succeeded', 'failed', 'cancelled') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'schema result is not terminal at required generation';
    END IF;
    job_id := v_job.job_id;
    result_generation := v_job.generation;
    succeeded := v_job.state = 'succeeded';
    IF succeeded THEN
        failure_code := NULL;
        failure_sha256 := NULL;
        receipt_sha256 := pg_catalog.sha256(pg_catalog.convert_to(
            pg_catalog.format(
                'forgecrud:v2:schema-result:v1|success|%s|%s|%s|%s',
                v_job.tenant_id, v_job.job_id, v_job.plan_sha256,
                v_job.generation
            ),
            'UTF8'
        ));
    ELSE
        failure_code := CASE v_job.state
            WHEN 'cancelled' THEN 'SCHEMA_JOB_CANCELLED'
            ELSE 'SCHEMA_JOB_FAILED'
        END;
        SELECT attempt.redacted_detail_sha256 INTO v_failure_digest
        FROM platform.schema_job_steps AS step
        JOIN platform.schema_job_step_attempts AS attempt
          ON attempt.job_id = step.job_id
         AND attempt.step_ordinal = step.ordinal
         AND attempt.attempt = step.attempt
         AND attempt.plan_sha256 = step.plan_sha256
         AND attempt.step_sha256 = step.step_sha256
        WHERE step.job_id = v_job.job_id
          AND step.ordinal = v_job.next_step_ordinal
          AND step.error_code = v_job.failure_code
          AND attempt.error_code = v_job.failure_code
          AND attempt.state IN ('failed', 'blocked', 'cancelled')
        FOR SHARE OF step, attempt;
        IF FOUND THEN
            failure_sha256 := pg_catalog.decode(v_failure_digest, 'hex');
        ELSIF v_job.state = 'cancelled' THEN
            failure_sha256 := pg_catalog.sha256(pg_catalog.convert_to(
                pg_catalog.format(
                    'forgecrud:v2:schema-result-cancelled:v1|%s|%s|%s|%s|%s',
                    v_job.tenant_id, v_job.job_id, v_job.plan_sha256,
                    v_job.generation, v_job.failure_code
                ),
                'UTF8'
            ));
        ELSE
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'schema terminal failure evidence differs';
        END IF;
        receipt_sha256 := pg_catalog.sha256(pg_catalog.convert_to(
            pg_catalog.format(
                'forgecrud:v2:schema-result:v1|failure|%s|%s|%s|%s|%s|%s',
                v_job.tenant_id, v_job.job_id, v_job.plan_sha256,
                v_job.generation, failure_code,
                pg_catalog.encode(failure_sha256, 'hex')
            ),
            'UTF8'
        ));
    END IF;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_resolve_schema_result_v1(uuid, uuid, text, bigint)
FROM PUBLIC;
