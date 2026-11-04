-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_receipt_read_v1(
    in_schema_name name,
    in_values uuid,
    in_values_2 uuid,
    in_kind text,
    in_values_3 bigint,
    in_values_4 text,
    in_values_5 bytea
)
RETURNS TABLE (
    found boolean,
    job_id uuid,
    run_at timestamptz,
    reference text,
    digest bytea,
    dedupe text,
    priority integer
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        false,
        NULL::uuid,
        NULL::timestamptz,
        NULL::text,
        ''::bytea,
        NULL::text,
        NULL::integer
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_receipt_read_v1(name, uuid, uuid, text, bigint, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_receipt_read_v1(name, uuid, uuid, text, bigint, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_receipt_read_v1(name, uuid, uuid, text, bigint, text, bytea);
