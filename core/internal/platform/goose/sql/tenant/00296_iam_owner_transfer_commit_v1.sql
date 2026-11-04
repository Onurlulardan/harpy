-- +goose Up
CREATE FUNCTION iam_owner_transfer_commit_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_actor_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_expected_platform_clock bigint,
    in_outbox_event_id uuid,
    in_event_sha256 bytea
)
RETURNS TABLE (
    operation_id uuid,
    predecessor_user_id uuid,
    successor_user_id uuid,
    predecessor_token_version bigint,
    predecessor_authorization_version bigint,
    successor_token_version bigint,
    successor_authorization_version bigint,
    owner_policy_revision bigint,
    platform_clock bigint,
    head_generation bigint,
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
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_owner_transfer_commit_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, uuid, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_owner_transfer_commit_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, uuid, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_owner_transfer_commit_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, uuid, bytea);
