-- +goose Up
CREATE TABLE platform.billing_v1_adoption_runs (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    cutover_lease_id uuid NOT NULL,
    cutover_evidence_generation bigint NOT NULL,
    cutover_lease_sha256 bytea NOT NULL,
    cutover_batch_sha256 bytea NOT NULL,
    cutover_session_seal_receipt_id uuid NOT NULL,
    cutover_session_seal_sha256 bytea NOT NULL,
    cutover_envelope_sha256 bytea NOT NULL,
    contract_version text NOT NULL,
    plan_sha256 bytea NOT NULL,
    source_sha256 bytea NOT NULL,
    plan_bytes bytea NOT NULL,
    plan_document jsonb NOT NULL,
    state text NOT NULL,
    eligible boolean NOT NULL,
    issue_count integer NOT NULL,
    snapshot_at timestamptz NOT NULL,
    source_timezone text NOT NULL,
    applied_count bigint,
    baseline_sha256 bytea,
    target_sha256 bytea,
    money_difference bigint,
    period_difference bigint,
    coupon_difference bigint,
    usage_difference bigint,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    applied_at timestamptz,
    verified_at timestamptz,
    rehearsed_at timestamptz,
    CONSTRAINT billing_v1_adoption_runs_pkey
        PRIMARY KEY (operation_id, run_id),
    CONSTRAINT billing_v1_adoption_runs_cutover_identity_check CHECK (
        cutover_lease_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND cutover_session_seal_receipt_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND cutover_evidence_generation > 0
    ),
    CONSTRAINT billing_v1_adoption_runs_scope_check CHECK (
        (scope_kind = 'global' AND scope_id <>
            '00000000-0000-0000-0000-000000000000'::uuid)
        OR (scope_kind = 'tenant' AND scope_id <>
            '00000000-0000-0000-0000-000000000000'::uuid)
    ),
    CONSTRAINT billing_v1_adoption_runs_contract_check CHECK (
        contract_version = 'forgecrud/billing-v1-adoption-plan/v1'
        AND source_timezone = 'UTC'
    ),
    CONSTRAINT billing_v1_adoption_runs_digest_check CHECK (
        pg_catalog.octet_length(plan_sha256) = 32
        AND pg_catalog.octet_length(source_sha256) = 32
        AND pg_catalog.octet_length(cutover_lease_sha256) = 32
        AND pg_catalog.octet_length(cutover_batch_sha256) = 32
        AND pg_catalog.octet_length(cutover_session_seal_sha256) = 32
        AND pg_catalog.octet_length(cutover_envelope_sha256) = 32
        AND (baseline_sha256 IS NULL OR
            pg_catalog.octet_length(baseline_sha256) = 32)
        AND (target_sha256 IS NULL OR
            pg_catalog.octet_length(target_sha256) = 32)
    ),
    CONSTRAINT billing_v1_adoption_runs_document_check CHECK (
        pg_catalog.octet_length(plan_bytes) BETWEEN 2 AND 67108864
        AND pg_catalog.jsonb_typeof(plan_document) = 'object'
        AND plan_document <> '{}'::jsonb
    ),
    CONSTRAINT billing_v1_adoption_runs_state_check CHECK (
        state IN (
            'blocked', 'planned', 'applied', 'verified',
            'rollback_rehearsed'
        )
    ),
    CONSTRAINT billing_v1_adoption_runs_issue_check CHECK (
        issue_count >= 0 AND eligible = (issue_count = 0)
        AND (eligible OR state = 'blocked')
    ),
    CONSTRAINT billing_v1_adoption_runs_metric_check CHECK (
        (applied_count IS NULL OR applied_count >= 0)
        AND (money_difference IS NULL OR money_difference >= 0)
        AND (period_difference IS NULL OR period_difference >= 0)
        AND (coupon_difference IS NULL OR coupon_difference >= 0)
        AND (usage_difference IS NULL OR usage_difference >= 0)
    ),
    CONSTRAINT billing_v1_adoption_runs_time_check CHECK (
        updated_at >= created_at
        AND (applied_at IS NULL OR applied_at BETWEEN created_at AND updated_at)
        AND (verified_at IS NULL OR verified_at BETWEEN created_at AND updated_at)
        AND (rehearsed_at IS NULL OR rehearsed_at BETWEEN created_at AND updated_at)
    )
);

CREATE INDEX billing_v1_adoption_runs_scope_idx
    ON platform.billing_v1_adoption_runs (
        scope_kind, scope_id, state, updated_at
    );
