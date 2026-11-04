-- +goose Up
CREATE FUNCTION platform.iam_recovery_anchor_commit_v1(
    in_operation_id uuid,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_now timestamptz
)
RETURNS TABLE (
    operation_id bytea,
    anchor_id bytea,
    kind text,
    authority_generation bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::bytea,
        ''::bytea,
        ''::text,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_commit_v1(uuid, uuid, uuid, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_commit_v1(uuid, uuid, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_anchor_commit_v1(uuid, uuid, uuid, timestamptz);
