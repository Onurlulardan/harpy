-- +goose Up
CREATE FUNCTION platform.iam_post_restore_challenge_context_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_restore_operation_id uuid,
    in_origin text,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    tenant_id uuid,
    owner_user_id uuid,
    anchor_id uuid,
    challenge text,
    credential_id bytea,
    public_key_cose bytea,
    rp_id text,
    sign_count bigint,
    context_sha256 bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::bytea,
        ''::bytea,
        ''::text,
        0::bigint,
        ''::bytea
    ;
$$;

REVOKE ALL ON FUNCTION platform.iam_post_restore_challenge_context_v1(uuid, uuid, uuid, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_post_restore_challenge_context_v1(uuid, uuid, uuid, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_post_restore_challenge_context_v1(uuid, uuid, uuid, text, timestamptz);
