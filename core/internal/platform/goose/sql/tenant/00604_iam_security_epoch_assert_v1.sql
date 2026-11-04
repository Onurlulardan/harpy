-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_security_epoch_assert_v1(
    in_schema_name name,
    in_epoch bytea
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_security_epoch_assert_v1(name, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_security_epoch_assert_v1(name, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_security_epoch_assert_v1(name, bytea);
