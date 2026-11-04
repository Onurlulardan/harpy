-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_initial_password_reject_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_selector bytea,
    in_secret_hmac bytea,
    in_csrf_hmac bytea,
    in_expected_revision bigint
)
RETURNS TABLE (
    state text,
    policy_attempts smallint,
    revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        0::smallint,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_password_reject_v1(name, uuid, bytea, bytea, bytea, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_initial_password_reject_v1(name, uuid, bytea, bytea, bytea, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_initial_password_reject_v1(name, uuid, bytea, bytea, bytea, bigint);
