-- +goose Up
CREATE TABLE platform.iam_v1_protected_authority_adoption_receipts (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL UNIQUE,
    primary_tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    authorization_version bigint NOT NULL CHECK (authorization_version > 0),
    subject_id uuid NOT NULL UNIQUE,
    owner_role_id uuid NOT NULL UNIQUE,
    owner_binding_id uuid NOT NULL UNIQUE,
    protected_mapping_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(protected_mapping_sha256)=32),
    plan_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(plan_sha256)=32),
    adopted_at timestamptz NOT NULL,
    recorded_by name NOT NULL CHECK (recorded_by='forgecrud_adopter'),
    CONSTRAINT iam_v1_protected_authority_mutation_fk
        FOREIGN KEY (operation_id)
        REFERENCES platform.iam_platform_mutation_receipts (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_v1_protected_authority_subject_fk
        FOREIGN KEY (environment_id,subject_id)
        REFERENCES platform.iam_platform_subjects (environment_id,id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_v1_protected_authority_owner_fk
        FOREIGN KEY (environment_id,owner_binding_id)
        REFERENCES platform.iam_platform_subject_roles (environment_id,id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TRIGGER iam_v1_protected_authority_receipts_guard
BEFORE UPDATE OR DELETE ON platform.iam_v1_protected_authority_adoption_receipts
FOR EACH ROW EXECUTE FUNCTION platform.iam_platform_receipt_guard_v1();
ALTER TABLE platform.iam_v1_protected_authority_adoption_receipts
    ENABLE ALWAYS TRIGGER iam_v1_protected_authority_receipts_guard;
