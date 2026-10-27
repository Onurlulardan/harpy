-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_bootstrap_v1(
    p_scope_schema name, p_tenant_id uuid, p_kind text,
    p_job_id uuid, p_run_at timestamptz
)
RETURNS TABLE (payload_reference text, payload_sha256 bytea, created boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_reference text;
    v_digest bytea;
    v_cadence integer;
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_created boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR p_job_id IS NULL OR
       pg_catalog.substring(p_job_id::text, 15, 1) <> '7' OR
       p_run_at IS NULL OR p_kind NOT IN (
           'iam.directory.repair', 'iam.directory.checksum',
           'iam.credential.bcrypt-retire', 'iam.identifier.retire',
           'iam.directory.tombstone-purge', 'iam.security-artifact-purge'
       ) OR NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema
       ) OR NOT EXISTS (
           SELECT 1 FROM platform.domain_cutovers AS cutover
           WHERE cutover.scope_kind = 'tenant' AND cutover.scope_id = p_tenant_id
             AND cutover.domain = 'iam' AND cutover.state = 'v2'
             AND cutover.writer = 'v2'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM maintenance bootstrap authority is invalid';
    END IF;
    v_cadence := CASE p_kind
        WHEN 'iam.directory.repair' THEN 900
        WHEN 'iam.security-artifact-purge' THEN 3600
        ELSE 86400
    END;
    v_reference := 'urn:forgecrud:iam:maintenance:' || p_tenant_id::text ||
        ':' || p_kind || ':v1';
    v_digest := pg_catalog.sha256(pg_catalog.convert_to(v_reference, 'UTF8'));
    INSERT INTO iam_maintenance_schedules (
        kind, payload_revision, payload_reference, payload_sha256,
        current_job_id, scheduled_run_at, cadence_seconds, revision,
        created_at, updated_at
    ) VALUES (
        p_kind, 1, v_reference, v_digest, p_job_id, p_run_at, v_cadence,
        1, v_now, v_now
    ) ON CONFLICT (kind) DO NOTHING;
    v_created := FOUND;
    IF NOT v_created AND NOT EXISTS (
        SELECT 1 FROM iam_maintenance_schedules AS schedule
        WHERE schedule.kind = p_kind AND schedule.payload_revision = 1
          AND schedule.payload_reference = v_reference
          AND schedule.payload_sha256 = v_digest
          AND schedule.current_job_id = p_job_id
          AND schedule.scheduled_run_at = p_run_at
          AND schedule.cadence_seconds = v_cadence
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'IAM maintenance schedule identity conflicts';
    END IF;
    INSERT INTO iam_maintenance_checkpoints (
        kind, cursor_id, scanned_count, affected_count, error_count,
        result_sha256, last_job_id, revision, updated_at
    ) VALUES (
        p_kind, NULL, 0, 0, 0, pg_catalog.sha256(''::bytea),
        NULL, 1, v_now
    ) ON CONFLICT (kind) DO NOTHING;
    RETURN QUERY SELECT v_reference, v_digest, v_created;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_bootstrap_v1(
    name,uuid,text,uuid,timestamptz
) FROM PUBLIC;
