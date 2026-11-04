-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_options_v1(
    in_operation_id uuid,
    in_anchor_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_idempotency_key text,
    in_request_sha256 bytea,
    in_challenge text,
    in_rpid text,
    in_origins jsonb,
    in_origins_sha256 bytea,
    in_expires_at timestamptz,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    operation_id bytea,
    anchor_id bytea,
    challenge text,
    rp_id text,
    allowed_origins bytea,
    expires_at timestamptz,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::bytea,
        ''::bytea,
        in_challenge,
        ''::text,
        ''::bytea,
        in_expires_at,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_options_v1(uuid, uuid, uuid, uuid, uuid, text, bytea, text, text, jsonb, bytea, timestamptz, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_options_v1(uuid, uuid, uuid, uuid, uuid, text, bytea, text, text, jsonb, bytea, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_anchor_options_v1(uuid, uuid, uuid, uuid, uuid, text, bytea, text, text, jsonb, bytea, timestamptz, timestamptz);
