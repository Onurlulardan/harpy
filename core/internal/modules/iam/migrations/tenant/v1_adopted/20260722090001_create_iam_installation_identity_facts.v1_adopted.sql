-- +goose Up
CREATE TABLE iam_installation_identity_facts (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    first_user_id uuid NOT NULL UNIQUE,
    root_organization_id uuid NOT NULL UNIQUE,
    protected_admin_binding_id uuid NOT NULL UNIQUE,
    protected_admin_binding_set_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(protected_admin_binding_set_sha256) = 32),
    normalized_email text COLLATE "C" NOT NULL,
    activated_at timestamptz NOT NULL
);

CREATE TRIGGER iam_installation_identity_facts_mutation_guard
BEFORE UPDATE OR DELETE ON iam_installation_identity_facts
FOR EACH ROW EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
