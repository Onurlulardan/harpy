-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_claim_assert_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_kind text, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR p_job_id IS NULL OR p_attempt_id IS NULL OR
       p_payload_revision IS DISTINCT FROM 1 OR
       pg_catalog.octet_length(p_payload_reference) NOT BETWEEN 1 AND 1024 OR
       pg_catalog.octet_length(p_payload_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.sha256(pg_catalog.convert_to(p_payload_reference, 'UTF8'))
           IS DISTINCT FROM p_payload_sha256 OR NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema
       ) OR NOT EXISTS (
           SELECT 1 FROM platform.domain_cutovers AS cutover
           WHERE cutover.scope_kind = 'tenant' AND cutover.scope_id = p_tenant_id
             AND cutover.domain = 'iam' AND cutover.state = 'v2'
             AND cutover.writer = 'v2'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM maintenance claim authority is invalid';
    END IF;
    IF p_kind = 'iam.profile.backfill' THEN
        IF NOT EXISTS (
            SELECT 1 FROM iam_profile_backfill_pages AS page
            WHERE page.job_id = p_job_id AND page.payload_reference = p_payload_reference
              AND page.payload_sha256 = p_payload_sha256 AND page.state = 'queued'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM profile backfill claim is stale';
        END IF;
    ELSIF NOT EXISTS (
        SELECT 1 FROM iam_maintenance_schedules AS schedule
        WHERE schedule.kind = p_kind AND schedule.payload_revision = p_payload_revision
          AND schedule.payload_reference = p_payload_reference
          AND schedule.payload_sha256 = p_payload_sha256
          AND schedule.current_job_id = p_job_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM recurring maintenance claim is stale';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_claim_assert_v1(
    name,uuid,uuid,uuid,text,integer,text,bytea
) FROM PUBLIC;
