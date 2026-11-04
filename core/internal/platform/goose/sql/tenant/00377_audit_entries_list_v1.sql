-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION audit_entries_list_v1(
    arg_1 text,
    arg_2 text,
    arg_3 text,
    arg_4 text,
    arg_5 text,
    arg_6 uuid,
    arg_7 uuid,
    arg_8 uuid,
    arg_9 timestamptz,
    arg_10 timestamptz,
    arg_11 text,
    arg_12 integer,
    arg_13 integer
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
    row_total bigint
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
        0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION audit_entries_list_v1(text, text, text, text, text, uuid, uuid, uuid, timestamptz, timestamptz, text, integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION audit_entries_list_v1(text, text, text, text, text, uuid, uuid, uuid, timestamptz, timestamptz, text, integer, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION audit_entries_list_v1(text, text, text, text, text, uuid, uuid, uuid, timestamptz, timestamptz, text, integer, integer);
