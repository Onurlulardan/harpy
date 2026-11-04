-- +goose Up
CREATE FUNCTION platform.tenancy_installation_delivery_record_v1(
    in_operation_id uuid,
    in_current_attempt_id uuid,
    in_current_credential_id uuid,
    in_current_credential_version bigint,
    in_next_attempt_id uuid,
    in_next_credential_id uuid,
    in_next_credential_version bigint,
    in_expected_revision bigint,
    in_recovery_token_sha256 text,
    in_manifest_sha256 text,
    in_cause text
)
RETURNS TABLE (
    delivery_state text,
    delivery_revision bigint,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        in_expected_revision,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_record_v1(uuid, uuid, uuid, bigint, uuid, uuid, bigint, bigint, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_record_v1(uuid, uuid, uuid, bigint, uuid, uuid, bigint, bigint, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_record_v1(uuid, uuid, uuid, bigint, uuid, uuid, bigint, bigint, text, text, text);
