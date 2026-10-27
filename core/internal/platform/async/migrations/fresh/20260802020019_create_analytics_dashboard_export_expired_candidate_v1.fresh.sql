-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_expired_candidate_v1()
RETURNS TABLE(
    tenant_id uuid,tenant_schema name,export_id uuid,job_id uuid,
    attempt_id uuid,attempt_generation bigint,lease_owner_id uuid,
    effect_started boolean,effect_id uuid,effect_request_sha256 bytea,
    cancellation_requested boolean,lease_expired_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export expired candidate authority is invalid';
    END IF;
    RETURN QUERY
    SELECT job.tenant_id,tenant.schema_name::name,job.payload_ref::uuid,job.id,
        attempt.attempt_id,attempt.attempt::bigint,attempt.lease_owner_id,
        attempt.effect_started_at IS NOT NULL,attempt.effect_id,
        CASE WHEN attempt.effect_request_sha256 IS NULL THEN NULL::bytea
             ELSE pg_catalog.decode(attempt.effect_request_sha256,'hex') END,
        job.cancel_requested_at IS NOT NULL,job.lease_until
    FROM platform.jobs AS job
    JOIN platform.job_attempts AS attempt
      ON attempt.job_id=job.id AND attempt.scope_kind='tenant'
     AND attempt.scope_id=job.tenant_id AND attempt.attempt=job.attempts
     AND attempt.lease_generation=job.lease_generation
     AND attempt.lease_owner_id=job.lease_owner_id
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
      AND job.state='running' AND job.lease_until<=pg_catalog.clock_timestamp()
      AND attempt.finished_at IS NULL
      AND attempt.attempt BETWEEN 1 AND 3
      AND pg_catalog.uuid_extract_version(attempt.attempt_id)=7
      AND pg_catalog.uuid_extract_version(attempt.lease_owner_id)=7
      AND ((attempt.effect_started_at IS NULL AND attempt.effect_id IS NULL AND
            attempt.effect_request_sha256 IS NULL AND job.effect_started_at IS NULL) OR
           (attempt.effect_started_at IS NOT NULL AND
            pg_catalog.uuid_extract_version(attempt.effect_id)=7 AND
            attempt.effect_request_sha256~'^[0-9a-f]{64}$' AND
            attempt.effect_request_sha256<>pg_catalog.repeat('0',64) AND
            job.effect_started_at=attempt.effect_started_at))
      AND tenant.deleted_at IS NULL
    ORDER BY job.lease_until,job.id LIMIT 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_expired_candidate_v1()
FROM PUBLIC;
