-- +goose Up
CREATE FUNCTION platform.platformops_complete_restore_execution_v2(
    in_operation_id text,
    in_attempt_id text,
    in_lease_token text,
    in_receipt_sha256 text,
    in_receipt_bytes bytea
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformops_complete_restore_execution_v2(text, text, text, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_complete_restore_execution_v2(text, text, text, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_complete_restore_execution_v2(text, text, text, text, bytea);
