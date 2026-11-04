-- +goose Up
CREATE FUNCTION platform.iam_maintenance_directory_apply_v1(
    in_arg uuid,
    in_id uuid,
    in_user_id uuid,
    in_kind text,
    in_state text,
    in_source_revision bigint,
    in_normalized_sha bytea,
    in_evidence jsonb
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.iam_maintenance_directory_apply_v1(uuid, uuid, uuid, text, text, bigint, bytea, jsonb) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_maintenance_directory_apply_v1(uuid, uuid, uuid, text, text, bigint, bytea, jsonb) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_maintenance_directory_apply_v1(uuid, uuid, uuid, text, text, bigint, bytea, jsonb);
