-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_schedules_v1(
    p_operation_id uuid,
    p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_kind text;
    v_cadence integer;
    v_reference text;
BEGIN
    FOR v_kind, v_cadence IN
        SELECT * FROM (VALUES
            ('iam.directory.repair', 900),
            ('iam.directory.checksum', 86400),
            ('iam.credential.bcrypt-retire', 86400),
            ('iam.identifier.retire', 86400),
            ('iam.directory.tombstone-purge', 86400),
            ('iam.security-artifact-purge', 3600)
        ) AS schedule(kind, cadence)
    LOOP
        v_reference := '{"contract":"iam-maintenance-schedule.v1","kind":"' ||
            v_kind || '"}';
        INSERT INTO iam_maintenance_schedules (
            kind, payload_revision, payload_reference, payload_sha256,
            current_job_id, scheduled_run_at, cadence_seconds, revision,
            created_at, updated_at
        ) VALUES (
            v_kind, 1, v_reference,
            pg_catalog.sha256(pg_catalog.convert_to(v_reference, 'UTF8')),
            pg_catalog.md5(p_operation_id::text || ':' || v_kind)::uuid,
            p_now + pg_catalog.make_interval(secs => v_cadence),
            v_cadence, 1, p_now, p_now
        );
    END LOOP;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_schedules_v1(
    uuid, timestamptz
) FROM PUBLIC;
