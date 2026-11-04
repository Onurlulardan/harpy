-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_receipts_v1(
    in_schema_name name,
    in_transfer_id uuid,
    in_actor_id uuid,
    in_resource text,
    in_token_version bigint,
    in_authorization_version bigint,
    in_after bigint,
    in_limit bigint
)
RETURNS TABLE (
    row_index bigint,
    outcome text,
    owner_kind text,
    owner_id uuid,
    error_code text,
    source_sha256 bytea,
    result_sha256 bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        ''::text,
        ''::text,
        NULL::uuid,
        ''::text,
        ''::bytea,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_receipts_v1(name, uuid, uuid, text, bigint, bigint, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_import_receipts_v1(name, uuid, uuid, text, bigint, bigint, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_import_receipts_v1(name, uuid, uuid, text, bigint, bigint, bigint, bigint);
