-- +goose Up
CREATE FUNCTION platform.platformops_ack_recovery_kit_v1(
    in_operation_id text,
    in_actor_user_id text,
    in_session_id text,
    in_artifact_sha256 bytea,
    in_ack_proof bytea,
    in_occurred_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformops_ack_recovery_kit_v1(text, text, text, bytea, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_ack_recovery_kit_v1(text, text, text, bytea, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_ack_recovery_kit_v1(text, text, text, bytea, bytea, timestamptz);
