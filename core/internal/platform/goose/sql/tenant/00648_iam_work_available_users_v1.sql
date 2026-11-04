-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_work_available_users_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_search text,
    in_limit integer
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT ''::bytea;
$$;

REVOKE ALL ON FUNCTION iam_work_available_users_v1(name, uuid, uuid, text, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_work_available_users_v1(name, uuid, uuid, text, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_work_available_users_v1(name, uuid, uuid, text, integer);
