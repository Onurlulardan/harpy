-- +goose Up
CREATE TABLE platform.cutover_billing_forms_rehearsals (
    receipt_id uuid PRIMARY KEY,
    scope_kind text NOT NULL,
    tenant_id uuid NOT NULL,
    group_id text NOT NULL,
    group_generation bigint NOT NULL,
    billing_operation_id uuid NOT NULL,
    billing_run_id uuid NOT NULL,
    billing_rollback_receipt_id uuid NOT NULL,
    billing_source_count bigint NOT NULL,
    billing_target_count bigint NOT NULL,
    billing_difference_count bigint NOT NULL,
    billing_acceptance_sha256 bytea NOT NULL,
    billing_rollback_sha256 bytea NOT NULL,
    forms_operation_id uuid NOT NULL,
    forms_run_id uuid NOT NULL,
    forms_rollback_receipt_id uuid NOT NULL,
    forms_source_count bigint NOT NULL,
    forms_target_count bigint NOT NULL,
    forms_difference_count bigint NOT NULL,
    forms_acceptance_sha256 bytea NOT NULL,
    forms_rollback_sha256 bytea NOT NULL,
    quiesce_drained boolean NOT NULL,
    usage_flushed boolean NOT NULL,
    production_dual_write boolean NOT NULL,
    group_sha256 bytea NOT NULL,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT cutover_billing_forms_rehearsals_group_fk
        FOREIGN KEY (scope_kind, tenant_id, group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT cutover_billing_forms_rehearsals_generation_key
        UNIQUE (scope_kind, tenant_id, group_id, group_generation),
    CONSTRAINT cutover_billing_forms_rehearsals_group_check CHECK (
        scope_kind = 'tenant' AND group_id = 'billing_forms' AND
        group_generation > 0
    ),
    CONSTRAINT cutover_billing_forms_rehearsals_count_check CHECK (
        billing_source_count >= 0 AND
        billing_source_count = billing_target_count AND
        billing_difference_count = 0 AND
        forms_source_count >= 0 AND
        forms_source_count = forms_target_count AND
        forms_difference_count = 0
    ),
    CONSTRAINT cutover_billing_forms_rehearsals_digest_check CHECK (
        pg_catalog.octet_length(billing_acceptance_sha256) = 32 AND
        pg_catalog.octet_length(billing_rollback_sha256) = 32 AND
        pg_catalog.octet_length(forms_acceptance_sha256) = 32 AND
        pg_catalog.octet_length(forms_rollback_sha256) = 32 AND
        pg_catalog.octet_length(group_sha256) = 32
    ),
    CONSTRAINT cutover_billing_forms_rehearsals_safety_check CHECK (
        quiesce_drained AND usage_flushed AND NOT production_dual_write
    )
);

CREATE TRIGGER cutover_billing_forms_rehearsals_immutable
BEFORE UPDATE OR DELETE ON platform.cutover_billing_forms_rehearsals
FOR EACH STATEMENT EXECUTE FUNCTION platform.cutover_reject_mutation();
ALTER TABLE platform.cutover_billing_forms_rehearsals ENABLE ALWAYS TRIGGER
    cutover_billing_forms_rehearsals_immutable;

