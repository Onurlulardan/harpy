-- +goose Up
CREATE TABLE platform.platform_cutover_v1_adoption_plans (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    deployment_mode text NOT NULL CHECK (deployment_mode IN ('onpremise','saas')),
    decision text NOT NULL CHECK (decision IN ('ready','fenced')),
    primary_tenant_id uuid,
    owner_user_id uuid,
    inventory_sha256 text NOT NULL CHECK (inventory_sha256 ~ '^[0-9a-f]{64}$'),
    plan_sha256 text CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    issue_codes jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(issue_codes)='array'),
    plan_document jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(plan_document)='object'),
    recorded_at timestamptz NOT NULL,
    recorded_by name NOT NULL CHECK (recorded_by='forgecrud_adopter'),
    CONSTRAINT platform_cutover_v1_adoption_decision_shape CHECK (
        (decision='ready' AND primary_tenant_id IS NOT NULL AND owner_user_id IS NOT NULL
            AND plan_sha256 IS NOT NULL AND pg_catalog.jsonb_array_length(issue_codes)=0)
        OR (decision='fenced' AND pg_catalog.jsonb_array_length(issue_codes)>0)
    )
);

CREATE TABLE platform.platform_cutover_v1_adoption_completion_receipts (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    frontend_contract_sha256 text NOT NULL
        CHECK (frontend_contract_sha256 ~ '^[0-9a-f]{64}$'),
    mcp_contract_sha256 text NOT NULL CHECK (mcp_contract_sha256 ~ '^[0-9a-f]{64}$'),
    legacy_boolean_hits bigint NOT NULL CHECK (legacy_boolean_hits=0),
    module_migration_hits bigint NOT NULL CHECK (module_migration_hits=0),
    zero_hit_since timestamptz NOT NULL,
    completed_at timestamptz NOT NULL,
    recorded_by name NOT NULL CHECK (recorded_by='forgecrud_adopter'),
    CONSTRAINT platform_cutover_v1_completion_plan_fk FOREIGN KEY (operation_id)
        REFERENCES platform.platform_cutover_v1_adoption_plans (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT platform_cutover_v1_completion_soak CHECK (
        zero_hit_since <= completed_at - interval '30 days'
    )
);

CREATE TRIGGER platform_cutover_v1_adoption_plans_reject_mutation
BEFORE UPDATE OR DELETE ON platform.platform_cutover_v1_adoption_plans
FOR EACH ROW EXECUTE FUNCTION platform.cutover_reject_mutation();
CREATE TRIGGER platform_cutover_v1_adoption_completion_reject_mutation
BEFORE UPDATE OR DELETE ON platform.platform_cutover_v1_adoption_completion_receipts
FOR EACH ROW EXECUTE FUNCTION platform.cutover_reject_mutation();
ALTER TABLE platform.platform_cutover_v1_adoption_plans
    ENABLE ALWAYS TRIGGER platform_cutover_v1_adoption_plans_reject_mutation;
ALTER TABLE platform.platform_cutover_v1_adoption_completion_receipts
    ENABLE ALWAYS TRIGGER platform_cutover_v1_adoption_completion_reject_mutation;
