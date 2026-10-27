-- +goose Up
CREATE TABLE platform.billing_v1_adoption_categories (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    category text NOT NULL,
    source_count bigint NOT NULL,
    source_sha256 bytea NOT NULL,
    target_count bigint,
    baseline_sha256 bytea,
    target_sha256 bytea,
    updated_at timestamptz NOT NULL,
    CONSTRAINT billing_v1_adoption_categories_pkey
        PRIMARY KEY (operation_id, run_id, category),
    CONSTRAINT billing_v1_adoption_categories_run_fk
        FOREIGN KEY (operation_id, run_id)
        REFERENCES platform.billing_v1_adoption_runs (operation_id, run_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_adoption_categories_name_check CHECK (
        category IN (
            'plans', 'campaigns', 'coupons', 'provider_accounts',
            'base_assignments', 'subscriptions', 'subscription_coupons',
            'submission_usage', 'gauge_openings', 'billing_events'
        )
    ),
    CONSTRAINT billing_v1_adoption_categories_count_check CHECK (
        source_count >= 0 AND (target_count IS NULL OR target_count >= 0)
    ),
    CONSTRAINT billing_v1_adoption_categories_digest_check CHECK (
        pg_catalog.octet_length(source_sha256) = 32
        AND (baseline_sha256 IS NULL OR
            pg_catalog.octet_length(baseline_sha256) = 32)
        AND (target_sha256 IS NULL OR
            pg_catalog.octet_length(target_sha256) = 32)
    )
);

CREATE TABLE platform.billing_v1_adoption_mappings (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    category text NOT NULL,
    source_key text COLLATE "C" NOT NULL,
    record_kind text NOT NULL,
    target_id uuid NOT NULL,
    record_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_v1_adoption_mappings_pkey
        PRIMARY KEY (operation_id, run_id, record_kind, source_key),
    CONSTRAINT billing_v1_adoption_mappings_target_key
        UNIQUE (operation_id, run_id, record_kind, target_id),
    CONSTRAINT billing_v1_adoption_mappings_category_fk
        FOREIGN KEY (operation_id, run_id, category)
        REFERENCES platform.billing_v1_adoption_categories (
            operation_id, run_id, category
        ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_adoption_mappings_key_check CHECK (
        pg_catalog.octet_length(source_key) BETWEEN 1 AND 512
        AND target_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT billing_v1_adoption_mappings_kind_check CHECK (
        record_kind IN (
            'offering', 'campaign', 'discount', 'provider_account',
            'base_assignment', 'subscription', 'redemption',
            'usage_opening', 'gauge_opening', 'legacy_event'
        )
    ),
    CONSTRAINT billing_v1_adoption_mappings_digest_check CHECK (
        pg_catalog.octet_length(record_sha256) = 32
    )
);

CREATE INDEX billing_v1_adoption_mappings_category_idx
    ON platform.billing_v1_adoption_mappings (
        operation_id, run_id, category, source_key
    );
