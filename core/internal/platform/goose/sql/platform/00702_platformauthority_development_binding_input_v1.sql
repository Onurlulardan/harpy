-- +goose Up
CREATE FUNCTION platform.platformauthority_development_binding_input_v1(
    in_environment_id text
)
RETURNS TABLE (
    operation_id text,
    registry_binding_id text,
    finalization_receipt_sha256 text,
    recovery_trust_bundle_sha256 text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        ''::text;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_development_binding_input_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_development_binding_input_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_development_binding_input_v1(text);
