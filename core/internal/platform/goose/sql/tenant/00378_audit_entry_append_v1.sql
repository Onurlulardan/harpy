-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION audit_entry_append_v1(
    arg_1 uuid,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 uuid,
    arg_6 text,
    arg_7 text,
    arg_8 text,
    arg_9 text,
    arg_10 text,
    arg_11 text,
    arg_12 bigint,
    arg_13 text,
    arg_14 jsonb,
    arg_15 bytea,
    arg_16 timestamptz,
    arg_17 timestamptz
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
    recorded_at timestamptz,
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
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
        pg_catalog.clock_timestamp(),
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION audit_entry_append_v1(uuid, uuid, uuid, uuid, uuid, text, text, text, text, text, text, bigint, text, jsonb, bytea, timestamptz, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION audit_entry_append_v1(uuid, uuid, uuid, uuid, uuid, text, text, text, text, text, text, bigint, text, jsonb, bytea, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION audit_entry_append_v1(uuid, uuid, uuid, uuid, uuid, text, text, text, text, text, text, bigint, text, jsonb, bytea, timestamptz, timestamptz);
