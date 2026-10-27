-- +goose Up
ALTER TABLE platform.billing_offering_entitlements
    ENABLE ALWAYS TRIGGER billing_offering_entitlements_authority_guard;
ALTER TABLE platform.billing_price_versions
    ENABLE ALWAYS TRIGGER billing_price_versions_authority_guard;
ALTER TABLE platform.billing_offering_versions
    ENABLE ALWAYS TRIGGER billing_offering_versions_authority_guard;
ALTER TABLE platform.billing_offerings
    ENABLE ALWAYS TRIGGER billing_offerings_authority_guard;

CREATE TABLE platform.billing_installation_base_assignments (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL UNIQUE,
    authority_id uuid NOT NULL UNIQUE,
    assignment_id uuid NOT NULL UNIQUE,
    deployment_mode text NOT NULL,
    offering_version_id uuid NOT NULL,
    offering_revision bigint NOT NULL,
    offering_sha256 text NOT NULL,
    offering_effective_at timestamptz NOT NULL,
    source_sequence bigint,
    source_sha256 text,
    manifest_sha256 text NOT NULL,
    assigned_at timestamptz NOT NULL,
    CONSTRAINT billing_installation_base_assignments_shape_check CHECK (
        deployment_mode IN ('onpremise','saas')
        AND offering_revision > 0
        AND offering_sha256 ~ '^[0-9a-f]{64}$'
        AND manifest_sha256 ~ '^[0-9a-f]{64}$'
        AND pg_catalog.isfinite(offering_effective_at)
        AND ((deployment_mode='onpremise' AND source_sequence > 0
              AND source_sha256 ~ '^[0-9a-f]{64}$')
          OR (deployment_mode='saas' AND source_sequence IS NULL
              AND source_sha256 IS NULL))
    )
);

CREATE TABLE platform.billing_installation_usage_openings (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL UNIQUE,
    meter_count integer NOT NULL CHECK (meter_count >= 0),
    opened_at timestamptz NOT NULL,
    CONSTRAINT billing_installation_usage_openings_operation_fk
        FOREIGN KEY (operation_id) REFERENCES
        platform.billing_installation_base_assignments (operation_id)
);

CREATE TRIGGER billing_installation_base_assignments_reject_mutation
BEFORE UPDATE OR DELETE ON platform.billing_installation_base_assignments
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
CREATE TRIGGER billing_installation_usage_openings_reject_mutation
BEFORE UPDATE OR DELETE ON platform.billing_installation_usage_openings
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
