-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_unknown_candidate_v1()
RETURNS TABLE(
    tenant_id uuid,tenant_schema name,export_id uuid,job_id uuid,
    attempt_id uuid,attempt_generation bigint,lease_owner_id uuid,
    effect_id uuid,effect_request_sha256 bytea,unknown_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export unknown candidate authority is invalid';
    END IF;
    RETURN QUERY
    SELECT job.tenant_id,tenant.schema_name::name,job.payload_ref::uuid,job.id,
        attempt.attempt_id,attempt.attempt::bigint,attempt.lease_owner_id,
        attempt.effect_id,pg_catalog.decode(attempt.effect_request_sha256,'hex'),
        job.completed_at
    FROM platform.jobs AS job
    JOIN platform.job_attempts AS attempt
      ON attempt.job_id=job.id AND attempt.scope_kind='tenant'
     AND attempt.scope_id=job.tenant_id AND attempt.attempt=job.attempts
     AND attempt.lease_generation=job.lease_generation
    JOIN platform.tenants AS tenant ON tenant.id=job.tenant_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind=job.scope_kind
     AND group_row.scope_id=job.scope_id
     AND group_row.group_id=job.authority_group_id
     AND group_row.state='v2' AND group_row.writer='v2'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=job.scope_kind
     AND domain_row.scope_id=job.scope_id
     AND domain_row.domain=job.authority_domain
     AND domain_row.state='v2' AND domain_row.writer='v2'
    JOIN platform.cutover_group_members AS member
      ON member.scope_kind=job.scope_kind AND member.scope_id=job.scope_id
     AND member.group_id=job.authority_group_id
     AND member.domain=job.authority_domain
    WHERE job.scope_kind='tenant' AND job.scope_id=job.tenant_id
      AND job.authority_domain='analytics' AND job.writer_origin='v2'
      AND job.kind='analytics.dashboard.export.render'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=3
      AND pg_catalog.pg_input_is_valid(job.payload_ref,'uuid')
      AND job.payload_ref=(job.payload_ref::uuid)::text
      AND pg_catalog.uuid_extract_version(job.payload_ref::uuid)=7
      AND job.payload_sha256=pg_catalog.encode(
          pg_catalog.sha256(pg_catalog.convert_to(job.payload_ref,'UTF8')),'hex')
      AND job.dedupe_key='analytics.dashboard.export.render:'||job.payload_ref
      AND job.effect_idempotency_key=job.dedupe_key
      AND job.state='dead' AND job.last_error_class IS NULL
      AND job.last_error_code='JOB_OUTCOME_UNKNOWN'
      AND job.last_error_sha256 IS NULL AND job.completed_at IS NOT NULL
      AND attempt.attempt BETWEEN 1 AND 3 AND attempt.outcome='dead'
      AND attempt.error_class IS NULL
      AND attempt.error_code='JOB_OUTCOME_UNKNOWN'
      AND attempt.redacted_error_sha256 IS NULL
      AND attempt.finished_at=job.completed_at
      AND attempt.effect_started_at IS NOT NULL
      AND attempt.effect_id IS NOT NULL
      AND pg_catalog.uuid_extract_version(attempt.lease_owner_id)=7
      AND attempt.effect_request_sha256~'^[0-9a-f]{64}$'
      AND attempt.effect_request_sha256<>pg_catalog.repeat('0',64)
      AND attempt.effect_receipt_id IS NULL
      AND tenant.deleted_at IS NULL
      AND NOT EXISTS (
        SELECT 1
        FROM platform.analytics_dashboard_export_unknown_resolutions resolution
        WHERE resolution.job_id=job.id AND resolution.attempt_id=attempt.attempt_id
          AND resolution.effect_id=attempt.effect_id
      )
      AND NOT EXISTS (
        SELECT 1 FROM platform.analytics_dashboard_export_unknown_cleanup_claims claim
        WHERE claim.job_id=job.id AND (
            claim.lease_until>pg_catalog.clock_timestamp() OR
            claim.run_at>pg_catalog.clock_timestamp())
      )
    ORDER BY job.completed_at,job.id LIMIT 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_unknown_candidate_v1()
FROM PUBLIC;
