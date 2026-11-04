-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_profile_read_contract_v2(
    in_schema_name name,
    in_resource text
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT ''::bytea;
$$;

REVOKE ALL ON FUNCTION iam_profile_read_contract_v2(name, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_profile_read_contract_v2(name, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_profile_read_contract_v2(name, text);
