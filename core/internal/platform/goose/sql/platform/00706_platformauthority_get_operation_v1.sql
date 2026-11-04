-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_get_operation_v1(
    in_operation_id text
)
RETURNS TABLE (
    operation_id text,
    purpose text,
    binding_id text,
    scope_kind text,
    scope_id text,
    generation bigint,
    previous_head text,
    state_sha text,
    state_bytes bytea,
    proposal_sha text,
    proposal_bytes bytea,
    compensates_ack text,
    status text,
    ack_sha text,
    ack_bytes bytea,
    registry_sequence bigint,
    completed_at timestamptz
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
        ''::text,
        ''::text,
        0::bigint,
        NULL::text,
        ''::text,
        ''::bytea,
        ''::text,
        ''::bytea,
        NULL::text,
        ''::text,
        NULL::text,
        ''::bytea,
        NULL::bigint,
        NULL::timestamptz;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_get_operation_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_get_operation_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_get_operation_v1(text);
