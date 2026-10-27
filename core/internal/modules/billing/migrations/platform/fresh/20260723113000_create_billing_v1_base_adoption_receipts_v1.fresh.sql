-- +goose Up
CREATE TABLE platform.billing_v1_base_adoption_receipts (
    operation_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL UNIQUE,
    assignment_id uuid NOT NULL UNIQUE,
    evidence_kind text NOT NULL CHECK (evidence_kind IN (
        'license_derived','saas_control','saas_free','saas_paid')),
    offering_key text NOT NULL,
    offering_version_id uuid NOT NULL,
    offering_revision bigint NOT NULL CHECK (offering_revision > 0),
    offering_sha256 text NOT NULL CHECK (offering_sha256 ~ '^[0-9a-f]{64}$'),
    offering_effective_at timestamptz NOT NULL,
    source_sequence bigint,
    source_sha256 text,
    subscription_evidence_sha256 text,
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    adopted_at timestamptz NOT NULL,
    recorded_by name NOT NULL CHECK (recorded_by='harpy_adopter'),
    CONSTRAINT billing_v1_base_adoption_pk PRIMARY KEY (operation_id,tenant_id),
    CONSTRAINT billing_v1_base_adoption_assignment_fk FOREIGN KEY (assignment_id)
        REFERENCES platform.billing_tenant_base_offerings (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_base_adoption_effective_at_finite CHECK (
        pg_catalog.isfinite(offering_effective_at)
    ),
    CONSTRAINT billing_v1_base_adoption_source_shape CHECK (
        (evidence_kind='license_derived' AND source_sequence>0
            AND source_sha256 ~ '^[0-9a-f]{64}$'
            AND subscription_evidence_sha256 IS NULL)
        OR (evidence_kind IN ('saas_control','saas_free')
            AND source_sequence IS NULL AND source_sha256 IS NULL
            AND subscription_evidence_sha256 IS NULL)
        OR (evidence_kind='saas_paid' AND source_sequence IS NULL
            AND source_sha256 IS NULL
            AND subscription_evidence_sha256 ~ '^[0-9a-f]{64}$')
    )
);

CREATE TRIGGER billing_v1_base_adoption_receipts_reject_mutation
BEFORE UPDATE OR DELETE ON platform.billing_v1_base_adoption_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_v1_base_adoption_receipts
    ENABLE ALWAYS TRIGGER billing_v1_base_adoption_receipts_reject_mutation;
