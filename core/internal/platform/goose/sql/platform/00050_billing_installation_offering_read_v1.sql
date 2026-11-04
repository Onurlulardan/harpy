-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_offering_read_v1(
    p_environment_id uuid,
    p_deployment_mode text
)
RETURNS TABLE (
    offering_version_id uuid,
    revision bigint,
    offering_sha256 text,
    effective_at timestamptz,
    source_sequence bigint,
    source_sha256 text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT offering.offering_version_id, offering.revision, offering.offering_sha256,
           offering.effective_at, offering.source_sequence, offering.source_sha256
    FROM platform.billing_license_offerings AS offering
    WHERE offering.environment_id = p_environment_id
      AND offering.deployment_mode = p_deployment_mode
    ORDER BY offering.revision DESC
    LIMIT 1;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_installation_offering_read_v1(uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_installation_offering_read_v1(uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_installation_offering_read_v1(uuid, text);
