-- +goose Up
CREATE TABLE platform.billing_license_offerings (
    offering_version_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    deployment_mode text,
    revision bigint NOT NULL,
    offering_sha256 text NOT NULL,
    effective_at timestamptz,
    source_sequence bigint,
    source_sha256 text
);

REVOKE ALL ON TABLE platform.billing_license_offerings FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON TABLE platform.billing_license_offerings TO harpy_api, harpy_worker;

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
    p_document bytea
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
    WITH existing AS (
        SELECT offering.offering_version_id, offering.revision, offering.offering_sha256
        FROM platform.billing_license_offerings AS offering
        WHERE offering.environment_id = p_environment_id
          AND offering.source_sequence IS NOT DISTINCT FROM p_sequence
          AND offering.source_sha256 IS NOT DISTINCT FROM p_snapshot_sha256
        ORDER BY offering.revision DESC
        LIMIT 1
    ), created AS (
        INSERT INTO platform.billing_license_offerings (
            offering_version_id, environment_id, deployment_mode, revision,
            offering_sha256, effective_at, source_sequence, source_sha256
        )
        SELECT
            COALESCE(p_offering_version_id, pg_catalog.gen_random_uuid()),
            p_environment_id, p_deployment_mode,
            COALESCE((
                SELECT pg_catalog.max(offering.revision)
                FROM platform.billing_license_offerings AS offering
                WHERE offering.environment_id = p_environment_id
            ), 0) + 1,
            CASE
                WHEN p_offering_sha256 ~ '^[0-9a-f]{64}$' THEN p_offering_sha256
                ELSE pg_catalog.md5(COALESCE(p_document, ''::bytea))
                    || pg_catalog.md5(COALESCE(p_snapshot_sha256, ''))
            END,
            p_effective_at, p_sequence, p_snapshot_sha256
        WHERE NOT EXISTS (SELECT 1 FROM existing)
        RETURNING offering_version_id, revision, offering_sha256
    )
    SELECT p_operation_id, created.offering_version_id, created.revision,
           created.offering_sha256, false
    FROM created
    UNION ALL
    SELECT p_operation_id, existing.offering_version_id, existing.revision,
           existing.offering_sha256, true
    FROM existing;
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
DROP TABLE platform.billing_license_offerings;
