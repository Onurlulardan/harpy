-- +goose Up
CREATE FUNCTION platform.iam_post_restore_challenge_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_restore_operation_id uuid,
    in_challenge text,
    in_context_sha256 bytea,
    in_rpid text,
    in_origin text,
    in_expires_at timestamptz,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    owner_user_id uuid,
    anchor_id uuid,
    challenge text,
    credential_id bytea,
    public_key_cose bytea,
    rp_id text,
    sign_count bigint,
    expires_at timestamptz,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_operation_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_challenge,
        ''::bytea,
        ''::bytea,
        ''::text,
        0::bigint,
        in_expires_at,
        false
    ;
$$;

REVOKE ALL ON FUNCTION platform.iam_post_restore_challenge_v1(uuid, uuid, uuid, text, bytea, text, text, timestamptz, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_post_restore_challenge_v1(uuid, uuid, uuid, text, bytea, text, text, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_post_restore_challenge_v1(uuid, uuid, uuid, text, bytea, text, text, timestamptz, timestamptz);
