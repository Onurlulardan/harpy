-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_enqueue_forms_schema_completion_v1(
    p_schema_job_id uuid,
    p_correlation_id uuid,
    p_authority_id uuid,
    p_authority_group_id text,
    p_authority_group_generation bigint,
    p_billing_generation bigint,
    p_forms_generation bigint
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_schema_job platform.schema_jobs%ROWTYPE;
    v_current_authority_id uuid;
    v_current_group_id text;
    v_current_group_generation bigint;
    v_current_billing_generation bigint;
    v_current_forms_generation bigint;
    v_payload_ref text;
    v_payload_sha256 text;
    v_dedupe_key text;
    v_now timestamptz;
    v_job_id uuid;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR SESSION_USER NOT IN ('forgecrud_migrator', 'forgecrud_v2_runtime')
       OR pg_catalog.uuid_extract_version(p_schema_job_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_correlation_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_authority_id) IS DISTINCT FROM 7
       OR p_authority_group_id IS DISTINCT FROM 'billing_forms'
       OR p_authority_group_generation NOT BETWEEN 1 AND 9007199254740991
       OR p_billing_generation NOT BETWEEN 1 AND 9007199254740991
       OR p_forms_generation NOT BETWEEN 1 AND 9007199254740991 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms schema completion authority is invalid';
    END IF;

    SELECT candidate.* INTO STRICT v_schema_job
    FROM platform.schema_jobs AS candidate
    WHERE candidate.job_id = p_schema_job_id
      AND candidate.owner = 'forms'
      AND candidate.resource_kind = 'form_storage'
      AND candidate.plan_contract_version = 'forms-schema-plan.v2'
      AND candidate.compiler_version = '2.0.0'
      AND candidate.state IN ('succeeded', 'failed', 'cancelled')
      AND candidate.finished_at IS NOT NULL
    FOR SHARE;

    IF SESSION_USER = 'forgecrud_migrator' THEN
        IF p_correlation_id IS DISTINCT FROM p_schema_job_id
           OR p_authority_id IS DISTINCT FROM
                v_schema_job.submission_authority_id
           OR p_authority_group_id IS DISTINCT FROM
                v_schema_job.submission_authority_group_id
           OR p_authority_group_generation IS DISTINCT FROM
                v_schema_job.submission_group_generation
           OR p_billing_generation IS DISTINCT FROM
                v_schema_job.submission_billing_generation
           OR p_forms_generation IS DISTINCT FROM
                v_schema_job.submission_forms_generation THEN
            RAISE EXCEPTION USING ERRCODE = '42501',
                MESSAGE = 'Forms schema completion frozen authority differs';
        END IF;
    ELSE
        IF p_correlation_id = p_schema_job_id THEN
            RAISE EXCEPTION USING ERRCODE = '42501',
                MESSAGE = 'Forms schema replay correlation is invalid';
        END IF;
        SELECT receipt.authority_id, receipt.authority_group_id,
               receipt.authority_group_generation, receipt.billing_generation,
               receipt.forms_generation
        INTO v_current_authority_id, v_current_group_id,
             v_current_group_generation, v_current_billing_generation,
             v_current_forms_generation
        FROM platform.billing_resolve_current_tenant_authority_receipt_v1(
            v_schema_job.tenant_id
        ) AS receipt;
        IF NOT FOUND
           OR p_authority_id IS DISTINCT FROM v_current_authority_id
           OR p_authority_group_id IS DISTINCT FROM v_current_group_id
           OR p_authority_group_generation IS DISTINCT FROM
                v_current_group_generation
           OR p_billing_generation IS DISTINCT FROM
                v_current_billing_generation
           OR p_forms_generation IS DISTINCT FROM v_current_forms_generation THEN
            RAISE EXCEPTION USING ERRCODE = '42501',
                MESSAGE = 'Forms schema replay authority differs';
        END IF;
    END IF;

    v_payload_ref :=
        'forms-publication-apply-result/' || p_schema_job_id::text;
    v_payload_sha256 := pg_catalog.encode(
        pg_catalog.sha256(pg_catalog.convert_to(v_payload_ref, 'UTF8')),
        'hex'
    );
    v_dedupe_key :=
        'forms.publication.apply-result:' || p_correlation_id::text;
    v_now := pg_catalog.clock_timestamp();
    v_job_id := pg_catalog.uuidv7();

    INSERT INTO platform.jobs (
        id, tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, cutover_generation, writer_origin,
        kind, payload_revision, payload_ref, payload_sha256,
        dedupe_key, effect_class, priority, run_at, state,
        attempts, max_attempts, lease_generation, created_at, updated_at
    ) VALUES (
        v_job_id, v_schema_job.tenant_id, 'tenant', v_schema_job.tenant_id,
        p_authority_group_id, p_authority_group_generation,
        'forms', p_forms_generation, 'v2',
        'forms.publication.apply-result', 1, v_payload_ref, v_payload_sha256,
        v_dedupe_key, 'database_transactional', 100, v_now, 'queued',
        0, 8, 0, v_now, v_now
    )
    ON CONFLICT ON CONSTRAINT jobs_dedupe_unique DO NOTHING
    RETURNING id INTO v_job_id;

    IF NOT FOUND THEN
        SELECT existing.id INTO STRICT v_job_id
        FROM platform.jobs AS existing
        WHERE existing.scope_kind = 'tenant'
          AND existing.scope_id = v_schema_job.tenant_id
          AND existing.tenant_id = v_schema_job.tenant_id
          AND existing.authority_group_id = p_authority_group_id
          AND existing.authority_group_generation =
                p_authority_group_generation
          AND existing.authority_domain = 'forms'
          AND existing.cutover_generation = p_forms_generation
          AND existing.writer_origin = 'v2'
          AND existing.kind = 'forms.publication.apply-result'
          AND existing.payload_revision = 1
          AND existing.payload_ref = v_payload_ref
          AND existing.payload_sha256 = v_payload_sha256
          AND existing.dedupe_key = v_dedupe_key
          AND existing.effect_class = 'database_transactional'
          AND existing.effect_idempotency_key IS NULL
          AND existing.max_attempts = 8
        FOR SHARE;
    END IF;
    RETURN v_job_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_async_enqueue_forms_schema_completion_v1(
    uuid,uuid,uuid,text,bigint,bigint,bigint
)
FROM PUBLIC;

