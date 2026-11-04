-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_profile_publication_receipt_read_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_actor_id uuid,
    in_manage_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    command_key text,
    request_sha256 bytea,
    result_id uuid,
    target_schema_id text,
    result_revision bigint,
    authorization_clock bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::bytea,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        0::bigint,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_profile_publication_receipt_read_v1(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_profile_publication_receipt_read_v1(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_profile_publication_receipt_read_v1(name, uuid, uuid, uuid, bigint, bigint);
