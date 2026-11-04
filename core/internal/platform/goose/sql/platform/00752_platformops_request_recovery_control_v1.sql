-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_request_recovery_control_v1(
    in_operation_id text,
    in_environment_id text,
    in_actor_user_id text,
    in_idempotency_key text,
    in_request_sha256 bytea,
    in_kind text,
    in_arg_7 text,
    in_arg_8 bytea,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    operation_id text,
    state text,
    kit_claim_state text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::text,
        ''::text;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_request_recovery_control_v1(text, text, text, text, bytea, text, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_request_recovery_control_v1(text, text, text, text, bytea, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_request_recovery_control_v1(text, text, text, text, bytea, text, text, bytea, timestamptz);
