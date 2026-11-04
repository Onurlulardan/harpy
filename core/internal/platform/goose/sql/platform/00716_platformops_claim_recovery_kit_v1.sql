-- +goose Up
CREATE FUNCTION platform.platformops_claim_recovery_kit_v1(
    in_operation_id text,
    in_actor_user_id text,
    in_session_id text,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    artifact_bytes bytea,
    artifact_sha256 text,
    expires_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::bytea,
        ''::text,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.platformops_claim_recovery_kit_v1(text, text, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_claim_recovery_kit_v1(text, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_claim_recovery_kit_v1(text, text, text, timestamptz);
