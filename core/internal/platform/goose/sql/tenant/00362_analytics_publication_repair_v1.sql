-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_publication_repair_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 bigint,
    arg_4 timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION analytics_publication_repair_v1(name, uuid, bigint, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_publication_repair_v1(name, uuid, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_publication_repair_v1(name, uuid, bigint, timestamptz);
