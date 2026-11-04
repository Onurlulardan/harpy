-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_get_head_v1(
    in_registry_binding_id text,
    in_scope_kind text,
    in_scope_id text
)
RETURNS TABLE (
    registry_binding_id text,
    scope_kind text,
    scope_id text,
    generation bigint,
    state_sha256 text,
    ack_sha256 text,
    registry_sequence bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_get_head_v1(text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_get_head_v1(text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_get_head_v1(text, text, text);
