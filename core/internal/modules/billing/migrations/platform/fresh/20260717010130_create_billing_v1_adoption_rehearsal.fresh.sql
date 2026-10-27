-- +goose Up
CREATE TABLE platform.billing_v1_adoption_opening_evidence (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    evidence_kind text NOT NULL,
    source_key text COLLATE "C" NOT NULL,
    tenant_id uuid NOT NULL,
    entitlement_key text COLLATE "C" NOT NULL,
    meter_period_id uuid NOT NULL,
    entry_id uuid,
    owner_fact_id uuid,
    period_start timestamptz NOT NULL,
    period_end timestamptz NOT NULL,
    opening_value bigint NOT NULL,
    evidence_sha256 bytea NOT NULL,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT billing_v1_adoption_opening_evidence_pkey
        PRIMARY KEY (operation_id, run_id, evidence_kind, source_key),
    CONSTRAINT billing_v1_adoption_opening_evidence_run_fk
        FOREIGN KEY (operation_id, run_id)
        REFERENCES platform.billing_v1_adoption_runs (operation_id, run_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_adoption_opening_evidence_kind_check CHECK (
        evidence_kind = 'owner_gauge'
            AND entitlement_key IN (
                'documents.storage_bytes',
                'iam.active_users'
            ) AND owner_fact_id IS NOT NULL
    ),
    CONSTRAINT billing_v1_adoption_opening_evidence_value_check CHECK (
        opening_value >= 0
        AND period_start < period_end
        AND ((opening_value = 0 AND entry_id IS NULL)
            OR (opening_value > 0 AND entry_id IS NOT NULL))
    ),
    CONSTRAINT billing_v1_adoption_opening_evidence_digest_check CHECK (
        pg_catalog.octet_length(evidence_sha256) = 32
    )
);

CREATE TABLE platform.billing_v1_adoption_rollback_rehearsals (
    receipt_id uuid NOT NULL,
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    observed_source_sha256 bytea NOT NULL,
    reverse_sha256 bytea NOT NULL,
    target_sha256 bytea NOT NULL,
    zero_difference boolean NOT NULL,
    rehearsed_at timestamptz NOT NULL,
    CONSTRAINT billing_v1_adoption_rollback_rehearsals_pkey
        PRIMARY KEY (receipt_id),
    CONSTRAINT billing_v1_adoption_rollback_rehearsals_operation_key
        UNIQUE (operation_id, run_id),
    CONSTRAINT billing_v1_adoption_rollback_rehearsals_run_fk
        FOREIGN KEY (operation_id, run_id)
        REFERENCES platform.billing_v1_adoption_runs (operation_id, run_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_adoption_rollback_rehearsals_digest_check CHECK (
        pg_catalog.octet_length(observed_source_sha256) = 32
        AND pg_catalog.octet_length(reverse_sha256) = 32
        AND pg_catalog.octet_length(target_sha256) = 32
        AND zero_difference =
            (observed_source_sha256 = reverse_sha256)
    )
);

