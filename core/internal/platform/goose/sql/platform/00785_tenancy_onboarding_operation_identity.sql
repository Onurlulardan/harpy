-- +goose Up
ALTER TABLE platform.tenancy_onboarding_operations
    ADD COLUMN IF NOT EXISTS environment_id uuid,
    ADD COLUMN IF NOT EXISTS idempotency_key text,
    ADD COLUMN IF NOT EXISTS request_sha256 text,
    ADD COLUMN IF NOT EXISTS manifest_sha256 text,
    ADD COLUMN IF NOT EXISTS schema_name text,
    ADD COLUMN IF NOT EXISTS binding_purpose text,
    ADD COLUMN IF NOT EXISTS root_organization_id uuid,
    ADD COLUMN IF NOT EXISTS protected_admin_binding_id uuid,
    ADD COLUMN IF NOT EXISTS instance_owner_binding_id uuid,
    ADD COLUMN IF NOT EXISTS offering_version_id uuid,
    ADD COLUMN IF NOT EXISTS offering_revision bigint,
    ADD COLUMN IF NOT EXISTS offering_sha256 text,
    ADD COLUMN IF NOT EXISTS offering_effective_at timestamptz,
    ADD COLUMN IF NOT EXISTS offering_source_sequence bigint,
    ADD COLUMN IF NOT EXISTS offering_source_sha256 text,
    ADD COLUMN IF NOT EXISTS recovery_generation bigint,
    ADD COLUMN IF NOT EXISTS recovery_trust_bundle_sha256 text,
    ADD COLUMN IF NOT EXISTS registry_binding_id uuid,
    ADD COLUMN IF NOT EXISTS authority_scope_kind text,
    ADD COLUMN IF NOT EXISTS authority_scope_id uuid,
    ADD COLUMN IF NOT EXISTS authority_generation bigint,
    ADD COLUMN IF NOT EXISTS authority_previous_head_sha256 text,
    ADD COLUMN IF NOT EXISTS authority_state_sha256 text,
    ADD COLUMN IF NOT EXISTS authority_proposal_sha256 text,
    ADD COLUMN IF NOT EXISTS created_at timestamptz;

ALTER TABLE platform.tenancy_onboarding_operations
    DROP CONSTRAINT IF EXISTS tenancy_onboarding_operations_schema_name_check;

ALTER TABLE platform.tenancy_onboarding_operations
    ADD CONSTRAINT tenancy_onboarding_operations_schema_name_check
    CHECK (schema_name IS NULL OR schema_name ~ '^tenant_[0-9a-f]{32}$') NOT VALID;

ALTER TABLE platform.tenancy_onboarding_operations
    VALIDATE CONSTRAINT tenancy_onboarding_operations_schema_name_check;

-- +goose Down
ALTER TABLE platform.tenancy_onboarding_operations
    DROP CONSTRAINT IF EXISTS tenancy_onboarding_operations_schema_name_check;

ALTER TABLE platform.tenancy_onboarding_operations
    DROP COLUMN IF EXISTS environment_id,
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS request_sha256,
    DROP COLUMN IF EXISTS manifest_sha256,
    DROP COLUMN IF EXISTS schema_name,
    DROP COLUMN IF EXISTS binding_purpose,
    DROP COLUMN IF EXISTS root_organization_id,
    DROP COLUMN IF EXISTS protected_admin_binding_id,
    DROP COLUMN IF EXISTS instance_owner_binding_id,
    DROP COLUMN IF EXISTS offering_version_id,
    DROP COLUMN IF EXISTS offering_revision,
    DROP COLUMN IF EXISTS offering_sha256,
    DROP COLUMN IF EXISTS offering_effective_at,
    DROP COLUMN IF EXISTS offering_source_sequence,
    DROP COLUMN IF EXISTS offering_source_sha256,
    DROP COLUMN IF EXISTS recovery_generation,
    DROP COLUMN IF EXISTS recovery_trust_bundle_sha256,
    DROP COLUMN IF EXISTS registry_binding_id,
    DROP COLUMN IF EXISTS authority_scope_kind,
    DROP COLUMN IF EXISTS authority_scope_id,
    DROP COLUMN IF EXISTS authority_generation,
    DROP COLUMN IF EXISTS authority_previous_head_sha256,
    DROP COLUMN IF EXISTS authority_state_sha256,
    DROP COLUMN IF EXISTS authority_proposal_sha256,
    DROP COLUMN IF EXISTS created_at;
