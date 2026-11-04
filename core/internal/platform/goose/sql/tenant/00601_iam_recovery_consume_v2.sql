-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_recovery_consume_v2(
    in_schema_name name,
    in_binding bytea,
    in_hmac_key_id text,
    in_code_hmac bytea,
    in_code_id uuid,
    in_expected_code_revision bigint,
    in_authenticator_id uuid,
    in_expected_authenticator_revision bigint,
    in_user_id uuid,
    in_expected_user_revision bigint,
    in_expected_token_version bigint
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT 0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_recovery_consume_v2(name, bytea, text, bytea, uuid, bigint, uuid, bigint, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_recovery_consume_v2(name, bytea, text, bytea, uuid, bigint, uuid, bigint, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_recovery_consume_v2(name, bytea, text, bytea, uuid, bigint, uuid, bigint, uuid, bigint, bigint);
