-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_external_effect_evidence_v1(
    p_tenant_id uuid,
    p_scope_kind text,
    p_scope_id uuid,
    p_authority_group_id text,
    p_authority_group_generation bigint,
    p_authority_domain text,
    p_cutover_generation bigint,
    p_claim_job_id uuid,
    p_claim_attempt_id uuid,
    p_claim_lease_generation bigint,
    p_claim_kind text,
    p_claim_payload_revision integer,
    p_claim_payload_reference text,
    p_claim_payload_sha256 text,
    p_claim_dedupe_prefix text,
    p_job_id uuid,
    p_expected_kind text,
    p_expected_payload_revision integer,
    p_expected_payload_reference text,
    p_payload_sha256 text,
    p_expected_dedupe_key text,
    p_expected_effect_key text,
    p_provider text,
    p_operation text,
    p_causation_id uuid
)
RETURNS TABLE (
    job_state text,
    job_attempts integer,
    effect_started boolean,
    receipt_status text,
    provider_reference text,
    error_code text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    job_row platform.jobs%ROWTYPE;
    attempt_row platform.job_attempts%ROWTYPE;
    receipt_row platform.idempotency_receipts%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR p_scope_kind NOT IN ('global','tenant')
       OR p_scope_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR p_authority_group_generation < 1 OR p_cutover_generation < 1
       OR p_claim_lease_generation < 1 OR p_claim_payload_revision < 1
       OR NOT (p_claim_kind OPERATOR(pg_catalog.~)
               '^[a-z][a-z0-9_.-]{2,127}$')
       OR pg_catalog.octet_length(p_claim_payload_reference) NOT BETWEEN 1 AND 2048
       OR NOT (p_claim_payload_sha256 OPERATOR(pg_catalog.~) '^[0-9a-f]{64}$')
       OR pg_catalog.octet_length(p_claim_dedupe_prefix) NOT BETWEEN 1 AND 512
       OR p_expected_payload_revision < 1
       OR NOT (p_expected_kind OPERATOR(pg_catalog.~)
               '^[a-z][a-z0-9_.-]{2,127}$')
       OR pg_catalog.octet_length(p_expected_payload_reference)
          NOT BETWEEN 1 AND 2048
       OR NOT (p_payload_sha256 OPERATOR(pg_catalog.~) '^[0-9a-f]{64}$')
       OR pg_catalog.octet_length(p_expected_dedupe_key) NOT BETWEEN 1 AND 512
       OR pg_catalog.octet_length(p_expected_effect_key) NOT BETWEEN 1 AND 512
       OR NOT (p_provider OPERATOR(pg_catalog.~) '^[a-z][a-z0-9_.-]{2,127}$')
       OR NOT (p_operation OPERATOR(pg_catalog.~) '^[a-z][a-z0-9_.-]{2,127}$')
       OR p_causation_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR (p_scope_kind = 'tenant' AND
           (p_tenant_id IS NULL OR p_tenant_id <> p_scope_id))
       OR (p_scope_kind = 'global' AND p_tenant_id IS NOT NULL) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'external effect evidence identity is invalid';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM platform.jobs AS claim_job
        JOIN platform.job_attempts AS claim_attempt
         ON claim_attempt.job_id = claim_job.id
         AND claim_attempt.attempt_id = p_claim_attempt_id
         AND claim_attempt.attempt = claim_job.attempts
         AND claim_attempt.lease_generation = p_claim_lease_generation
        WHERE claim_job.id = p_claim_job_id
          AND claim_job.tenant_id IS NOT DISTINCT FROM p_tenant_id
          AND claim_job.scope_kind = p_scope_kind
          AND claim_job.scope_id = p_scope_id
          AND claim_job.authority_group_id = p_authority_group_id
          AND claim_job.authority_group_generation =
              p_authority_group_generation
          AND claim_job.authority_domain = p_authority_domain
          AND claim_job.cutover_generation = p_cutover_generation
          AND claim_job.writer_origin = 'v2'
          AND claim_job.kind = p_claim_kind
          AND claim_job.payload_revision = p_claim_payload_revision
          AND claim_job.lease_generation = p_claim_lease_generation
          AND claim_job.effect_class = 'database_transactional'
          AND claim_job.state = 'running'
          AND claim_job.payload_ref = p_claim_payload_reference
          AND claim_job.payload_sha256 = p_claim_payload_sha256
          AND (claim_job.dedupe_key = p_claim_dedupe_prefix
               OR pg_catalog.left(claim_job.dedupe_key,
                  pg_catalog.char_length(p_claim_dedupe_prefix) + 1) =
                  pg_catalog.concat(p_claim_dedupe_prefix, ':'))
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'database job claim evidence is invalid';
    END IF;

    SELECT * INTO job_row
    FROM platform.jobs AS job
    WHERE job.id = p_job_id
      AND job.tenant_id IS NOT DISTINCT FROM p_tenant_id
      AND job.scope_kind = p_scope_kind
      AND job.scope_id = p_scope_id
      AND job.authority_group_id = p_authority_group_id
      AND job.authority_group_generation = p_authority_group_generation
      AND job.authority_domain = p_authority_domain
      AND job.cutover_generation = p_cutover_generation
      AND job.writer_origin = 'v2'
      AND job.kind = p_expected_kind
      AND job.payload_revision = p_expected_payload_revision
      AND job.payload_ref = p_expected_payload_reference
      AND job.payload_sha256 = p_payload_sha256
      AND job.dedupe_key = p_expected_dedupe_key
      AND job.effect_class = 'external'
      AND job.effect_idempotency_key = p_expected_effect_key
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'external effect job authority is invalid';
    END IF;

    IF job_row.attempts > 0 THEN
        SELECT * INTO attempt_row
        FROM platform.job_attempts AS attempt
        WHERE attempt.job_id = job_row.id
          AND attempt.scope_kind = p_scope_kind
          AND attempt.scope_id = p_scope_id
          AND attempt.attempt = job_row.attempts
        FOR SHARE;
        IF attempt_row.attempt_id IS NULL OR
           (job_row.state NOT IN ('queued','running') AND attempt_row.outcome IS NULL) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5721',
                MESSAGE = 'external effect terminal attempt is invalid';
        END IF;
    ELSIF job_row.state NOT IN ('queued','running') THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'external effect terminal job has no attempt';
    END IF;
    IF attempt_row.effect_receipt_id IS NOT NULL THEN
        SELECT * INTO receipt_row
        FROM platform.idempotency_receipts AS receipt
        WHERE receipt.receipt_id = attempt_row.effect_receipt_id
        FOR SHARE;
        IF receipt_row.receipt_id IS NULL OR receipt_row.record_kind <> 'effect'
           OR receipt_row.tenant_id IS DISTINCT FROM p_tenant_id
           OR receipt_row.scope_kind <> p_scope_kind
           OR receipt_row.scope_id <> p_scope_id
           OR receipt_row.authority_group_id <> p_authority_group_id
           OR receipt_row.authority_group_generation <>
              p_authority_group_generation
           OR receipt_row.authority_domain <> p_authority_domain
           OR receipt_row.cutover_generation <> p_cutover_generation
           OR receipt_row.writer_origin <> 'v2'
           OR receipt_row.request_sha256 <> attempt_row.effect_request_sha256
           OR receipt_row.effect_id <> attempt_row.effect_id
           OR receipt_row.effect_kind <> job_row.kind
           OR receipt_row.effect_revision <> 1
           OR receipt_row.effect_key <> job_row.effect_idempotency_key
           OR receipt_row.provider <> p_provider
           OR receipt_row.operation <> p_operation
           OR receipt_row.causation_id <> p_causation_id THEN
            RAISE EXCEPTION USING ERRCODE = 'P5721',
                MESSAGE = 'external effect receipt chain is invalid';
        END IF;
    END IF;
    RETURN QUERY SELECT job_row.state, job_row.attempts,
        job_row.effect_started_at IS NOT NULL, receipt_row.effect_status,
        receipt_row.provider_reference,
        COALESCE(receipt_row.effect_code, job_row.last_error_code);
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_external_effect_evidence_v1(
    uuid,text,uuid,text,bigint,text,bigint,uuid,uuid,bigint,text,integer,
    text,text,text,uuid,text,integer,text,text,text,text,text,text,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_external_effect_evidence_v1(
    uuid,text,uuid,text,bigint,text,bigint,uuid,uuid,bigint,text,integer,
    text,text,text,uuid,text,integer,text,text,text,text,text,text,uuid
) TO forgecrud_runtime;
-- +goose StatementEnd
