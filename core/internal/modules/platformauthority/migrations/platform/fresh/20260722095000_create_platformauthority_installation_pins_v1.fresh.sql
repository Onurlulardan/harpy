-- +goose Up
CREATE TABLE platform.platformauthority_installation_owner_bindings (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL UNIQUE,
    tenant_id uuid NOT NULL UNIQUE,
    instance_owner_binding_id uuid NOT NULL UNIQUE,
    protected_admin_binding_id uuid NOT NULL,
    registry_binding_id uuid NOT NULL,
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL
);

CREATE TABLE platform.platformauthority_installation_head_pins (
    operation_id uuid PRIMARY KEY,
    registry_binding_id uuid NOT NULL,
    scope_kind text NOT NULL CHECK (scope_kind = 'environment'),
    scope_id uuid NOT NULL UNIQUE,
    generation bigint NOT NULL CHECK (generation > 0),
    previous_head_sha256 text,
    state_sha256 text NOT NULL CHECK (state_sha256 ~ '^[0-9a-f]{64}$'),
    proposal_sha256 text NOT NULL CHECK (proposal_sha256 ~ '^[0-9a-f]{64}$'),
    ack_sha256 text NOT NULL UNIQUE CHECK (ack_sha256 ~ '^[0-9a-f]{64}$'),
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    pinned_at timestamptz NOT NULL,
    CHECK ((generation = 1) = (previous_head_sha256 IS NULL))
);

CREATE TRIGGER platformauthority_installation_owner_binding_immutable
BEFORE UPDATE OR DELETE ON platform.platformauthority_installation_owner_bindings
FOR EACH ROW EXECUTE FUNCTION platform.platformauthority_reject_mutation_v1();

CREATE TRIGGER platformauthority_installation_head_pin_immutable
BEFORE UPDATE OR DELETE ON platform.platformauthority_installation_head_pins
FOR EACH ROW EXECUTE FUNCTION platform.platformauthority_reject_mutation_v1();
