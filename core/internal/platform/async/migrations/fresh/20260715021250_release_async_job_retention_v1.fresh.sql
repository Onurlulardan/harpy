-- +goose Up
ALTER TABLE platform.jobs DROP CONSTRAINT jobs_time_check;
ALTER TABLE platform.jobs ADD CONSTRAINT jobs_time_check CHECK (
    created_at <= updated_at AND
    (payload_redacted_at IS NULL OR completed_at <= payload_redacted_at) AND
    (payload_redacted_at IS NULL OR state IN ('succeeded', 'dead', 'cancelled')) AND
    (cancel_requested_at IS NULL OR created_at <= cancel_requested_at) AND
    (parked_at IS NULL OR created_at <= parked_at) AND
    (effect_started_at IS NULL OR created_at <= effect_started_at) AND
    (completed_at IS NULL OR completed_at <= updated_at) AND
    (retention_released_at IS NULL OR completed_at <= retention_released_at)
);

-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_release_job_retention_v1(
    p_job_id uuid,
    p_scope_id uuid,
    p_group_id text,
    p_group_generation bigint,
    p_owner_domain text,
    p_cutover_generation bigint,
    p_expected_revision bigint,
    p_release_sha256 text,
    p_retain_until timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_now timestamptz;
    v_released boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_job_id IS NULL OR p_scope_id IS NULL OR
       p_expected_revision IS NULL OR p_expected_revision < 0 OR
       p_release_sha256 IS NULL OR
       p_release_sha256 !~ '^[0-9a-f]{64}$' OR
       p_retain_until IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'job retention release input is invalid';
    END IF;

    IF platform.platform_async_assert_current_authority_v1(
        'tenant', p_scope_id, p_group_id, p_group_generation,
        p_owner_domain, p_cutover_generation
    ) IS NULL THEN NULL; END IF;

    v_now := pg_catalog.clock_timestamp();
    UPDATE platform.jobs AS job
    SET retain_until = p_retain_until,
        retention_release_revision = p_expected_revision + 1,
        retention_release_sha256 = p_release_sha256,
        retention_released_at = v_now,
        updated_at = v_now
    WHERE job.id = p_job_id
      AND job.scope_kind = 'tenant'
      AND job.scope_id = p_scope_id
      AND job.tenant_id = p_scope_id
      AND job.authority_group_id = p_group_id
      AND job.authority_domain = p_owner_domain
      AND job.state IN ('succeeded', 'dead', 'cancelled')
      AND job.retention_release_revision = p_expected_revision
      AND job.retain_until IS NOT NULL
      AND job.retain_until <= v_now
      AND p_retain_until >= job.retain_until
      AND p_retain_until <= v_now
      AND NOT EXISTS (
          SELECT 1 FROM platform.idempotency_receipts AS dependency
          WHERE dependency.async_job_id = job.id
            AND dependency.retain_until > v_now
      )
      AND NOT EXISTS (
          SELECT 1 FROM platform.jobs AS retry
          WHERE retry.retry_of_job_id = job.id
      )
      AND NOT EXISTS (
          SELECT 1 FROM platform.job_attempts AS attempt
          WHERE attempt.job_id = job.id AND attempt.finished_at IS NULL
      )
      AND NOT EXISTS (
          SELECT 1
          FROM platform.job_attempts AS attempt
          JOIN platform.idempotency_receipts AS receipt
            ON receipt.receipt_id = attempt.effect_receipt_id
          WHERE attempt.job_id = job.id AND receipt.retain_until > v_now
      )
    RETURNING true INTO v_released;

    RETURN COALESCE(v_released, false);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_release_job_retention_v1(
    uuid,uuid,text,bigint,text,bigint,bigint,text,timestamptz
) FROM PUBLIC;
