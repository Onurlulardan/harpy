-- +goose Up
CREATE TABLE platform.tenancy_v1_adoption_receipts (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL UNIQUE,
    deployment_mode text NOT NULL CHECK (deployment_mode IN ('onpremise','saas')),
    primary_tenant_id uuid NOT NULL UNIQUE,
    adopted_tenant_count integer NOT NULL CHECK (adopted_tenant_count > 0),
    inventory_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(inventory_sha256)=32),
    plan_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(plan_sha256)=32),
    adopted_at timestamptz NOT NULL,
    recorded_by name NOT NULL CHECK (recorded_by='harpy_adopter'),
    CONSTRAINT tenancy_v1_adoption_policy_fk FOREIGN KEY (environment_id)
        REFERENCES platform.tenancy_deployment_policy (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT tenancy_v1_adoption_primary_fk
        FOREIGN KEY (environment_id,primary_tenant_id)
        REFERENCES platform.tenancy_tenant_bindings (environment_id,tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TRIGGER tenancy_v1_adoption_receipts_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenancy_v1_adoption_receipts
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();
ALTER TABLE platform.tenancy_v1_adoption_receipts
    ENABLE ALWAYS TRIGGER tenancy_v1_adoption_receipts_reject_mutation;
