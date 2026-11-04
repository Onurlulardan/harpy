-- +goose Up
-- No throttle row exists, so the probe is not blocked.
CREATE FUNCTION platform.iam_auth_throttle_blocked_v2(
    p_scope_kind text,
    p_tenant_id uuid,
    p_purpose text,
    p_dimension text,
    p_key_id text,
    p_bucket_hmac bytea,
    p_group_generation bigint,
    p_iam_generation bigint
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.iam_auth_throttle_blocked_v2(text, uuid, text, text, text, bytea, bigint, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_auth_throttle_blocked_v2(text, uuid, text, text, text, bytea, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_auth_throttle_blocked_v2(text, uuid, text, text, text, bytea, bigint, bigint);
