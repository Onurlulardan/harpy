-- +goose Up
CREATE TABLE platform.tenancy_activation_audits (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL UNIQUE,
    actor_id uuid NOT NULL,
    event_type text NOT NULL CHECK (event_type = 'installation.activated'),
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    recorded_at timestamptz NOT NULL
);

INSERT INTO platform.tenancy_activation_audits (
    operation_id, environment_id, tenant_id, actor_id,
    event_type, manifest_sha256, recorded_at
)
SELECT fact.operation_id, fact.environment_id, fact.tenant_id,
       fact.owner_user_id, 'installation.activated',
       fact.manifest_sha256, fact.activated_at
FROM platform.tenancy_activation_facts AS fact;

CREATE TRIGGER tenancy_activation_audits_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenancy_activation_audits
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();
