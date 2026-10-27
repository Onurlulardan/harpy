-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_candidate_v1()
RETURNS TABLE(
    tenant_id uuid,
    tenant_schema name,
    export_id uuid,
    job_id uuid,
    attempt_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Analytics dashboard export candidate authority is invalid';
    END IF;
    RETURN QUERY
    SELECT job.tenant_id,tenant.schema_name::name,job.payload_ref::uuid,
           job.id,(job.attempts + 1)::bigint
    FROM platform.jobs AS job
    JOIN platform.tenants AS tenant ON tenant.id=job.tenant_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind=job.scope_kind
     AND group_row.scope_id=job.scope_id
     AND group_row.group_id=job.authority_group_id
     AND group_row.generation=job.authority_group_generation
     AND group_row.state='v2' AND group_row.writer='v2'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=job.scope_kind
     AND domain_row.scope_id=job.scope_id
     AND domain_row.domain=job.authority_domain
     AND domain_row.generation=job.cutover_generation
     AND domain_row.state='v2' AND domain_row.writer='v2'
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
      AND job.state='queued' AND job.parked_at IS NULL
      AND job.effect_started_at IS NULL
      AND job.attempts<job.max_attempts
      AND job.run_at<=pg_catalog.clock_timestamp()
      AND tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE','TRIAL')
    ORDER BY job.priority DESC,job.run_at,job.id
    LIMIT 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_candidate_v1()
FROM PUBLIC;
