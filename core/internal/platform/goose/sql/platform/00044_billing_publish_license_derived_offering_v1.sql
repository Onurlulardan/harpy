-- +goose Up
-- BELIRSIZ: offering persistence and replay are not determined by the caller.
CREATE FUNCTION platform.billing_publish_license_derived_offering_v1(
    p_operation_id uuid,
    p_environment_id uuid,
    p_license_id uuid,
    p_licensee_organization_id uuid,
    p_deployment_mode text,
    p_delivery_kind text,
    p_issuer text,
    p_sequence bigint,
    p_snapshot_sha256 text,
    p_offering_version_id uuid,
    p_revision bigint,
    p_offering_sha256 text,
    p_effective_at timestamptz,
    p_manifest_sha256 text,
    p_offering_document bytea
)
RETURNS TABLE (
    operation_id uuid,
    offering_version_id uuid,
    revision bigint,
    offering_sha256 text,
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        p_operation_id,
        p_offering_version_id,
        p_revision,
        p_offering_sha256,
        false;
$$;

REVOKE ALL ON FUNCTION platform.billing_publish_license_derived_offering_v1(
    uuid, uuid, uuid, uuid, text, text, text, bigint, text, uuid, bigint, text,
    timestamptz, text, bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_publish_license_derived_offering_v1(
    uuid, uuid, uuid, uuid, text, text, text, bigint, text, uuid, bigint, text,
    timestamptz, text, bytea
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_publish_license_derived_offering_v1(
    uuid, uuid, uuid, uuid, text, text, text, bigint, text, uuid, bigint, text,
    timestamptz, text, bytea
);
