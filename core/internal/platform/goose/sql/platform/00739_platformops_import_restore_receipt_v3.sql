-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_import_restore_receipt_v3(
    in_operation_id text,
    in_recovery_set_id text,
    in_environment_id text,
    in_plan_sha256 text,
    in_manifest_sha256 text,
    in_receipt_sha256 text,
    in_receipt_bytes bytea,
    in_previous_database_system_id text,
    in_new_database_system_id text,
    in_object_merkle_sha256 text,
    in_locator_remap_sha256 text,
    in_safety_evidence_sha256 text,
    in_drain_ack_sha256 text,
    in_switch_ack_sha256 text,
    in_control_database_locator_sha256 text,
    in_scratch_database_locator_sha256 text,
    in_component_inventory jsonb
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

REVOKE ALL ON FUNCTION platform.platformops_import_restore_receipt_v3(text, text, text, text, text, text, bytea, text, text, text, text, text, text, text, text, text, jsonb) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_import_restore_receipt_v3(text, text, text, text, text, text, bytea, text, text, text, text, text, text, text, text, text, jsonb) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_import_restore_receipt_v3(text, text, text, text, text, text, bytea, text, text, text, text, text, text, text, text, text, jsonb);
