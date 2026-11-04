-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION audit_entry_get_v1(
    arg_1 uuid
)
RETURNS TABLE (
    id uuid,
    owner_audit_id uuid,
    actor_id uuid,
    organization_id uuid,
    correlation_id uuid,
    owner_module text,
    operation text,
    entity_type text,
    entity_id text,
    form_slug text,
    action text,
    authorization_version bigint,
    reason_code text,
    field_changes bytea,
    event_sha256 bytea,
    occurred_at timestamptz,
    recorded_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        NULL::uuid,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        NULL::text,
        ''::text,
        0::bigint,
        NULL::text,
        NULL::bytea,
        NULL::bytea,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION audit_entry_get_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION audit_entry_get_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION audit_entry_get_v1(uuid);
