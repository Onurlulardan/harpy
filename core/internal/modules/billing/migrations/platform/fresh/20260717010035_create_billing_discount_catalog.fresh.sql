-- +goose Up
CREATE TABLE platform.billing_discounts (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    normalized_code text COLLATE "C" NOT NULL,
    lifecycle text NOT NULL,
    created_at timestamptz NOT NULL,
    revoked_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_discounts_pkey PRIMARY KEY (id),
    CONSTRAINT billing_discounts_code_key UNIQUE (normalized_code),
    CONSTRAINT billing_discounts_authority_id_key UNIQUE (authority_id, id),
    CONSTRAINT billing_discounts_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discounts_code_check CHECK (
        normalized_code ~ '^[A-Z0-9][A-Z0-9_-]{0,63}$'
    ),
    CONSTRAINT billing_discounts_lifecycle_check CHECK (
        (lifecycle = 'active' AND revoked_at IS NULL)
        OR (lifecycle = 'revoked' AND revoked_at IS NOT NULL)
    ),
    CONSTRAINT billing_discounts_revision_check CHECK (revision > 0),
    CONSTRAINT billing_discounts_time_check CHECK (
        revoked_at IS NULL OR revoked_at >= created_at
    )
);

CREATE TABLE platform.billing_discount_versions (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    discount_id uuid NOT NULL,
    version bigint NOT NULL,
    discount_kind text NOT NULL,
    fixed_minor_units bigint,
    currency character(3) COLLATE "C",
    currency_registry_version bigint,
    currency_eligible boolean,
    basis_points integer,
    duration text NOT NULL,
    repeating_cycles bigint,
    priority integer NOT NULL,
    global_use_limit bigint,
    per_tenant_use_limit bigint,
    all_offerings boolean NOT NULL,
    expected_offering_count smallint NOT NULL,
    publication_operation_id uuid NOT NULL,
    valid_start timestamptz NOT NULL,
    valid_end timestamptz,
    published_at timestamptz NOT NULL,
    CONSTRAINT billing_discount_versions_pkey PRIMARY KEY (id),
    CONSTRAINT billing_discount_versions_discount_version_key
        UNIQUE (discount_id, version),
    CONSTRAINT billing_discount_versions_operation_key
        UNIQUE (publication_operation_id),
    CONSTRAINT billing_discount_versions_authority_id_key UNIQUE (authority_id, id),
    CONSTRAINT billing_discount_versions_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_versions_discount_fk FOREIGN KEY (
        authority_id, discount_id
    ) REFERENCES platform.billing_discounts (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_versions_currency_fk FOREIGN KEY (
        currency_registry_version, currency, currency_eligible
    ) REFERENCES platform.billing_currency_exponents (
        registry_version, code, financial_money_eligible
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_versions_amount_check CHECK (
        (discount_kind = 'fixed' AND fixed_minor_units BETWEEN 1 AND 9000000000000000
            AND currency IS NOT NULL AND currency_registry_version > 0
            AND currency_eligible AND basis_points IS NULL)
        OR (discount_kind = 'basis_point' AND basis_points BETWEEN 1 AND 10000
            AND fixed_minor_units IS NULL AND currency IS NULL
            AND currency_registry_version IS NULL AND currency_eligible IS NULL)
    ),
    CONSTRAINT billing_discount_versions_duration_check CHECK (
        (duration = 'once' AND repeating_cycles IS NULL)
        OR (duration = 'repeating' AND repeating_cycles > 0)
        OR (duration = 'forever' AND repeating_cycles IS NULL)
    ),
    CONSTRAINT billing_discount_versions_policy_check CHECK (
        publication_operation_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND version > 0 AND priority >= 0
        AND (global_use_limit IS NULL OR global_use_limit > 0)
        AND (per_tenant_use_limit IS NULL OR per_tenant_use_limit > 0)
        AND ((all_offerings AND expected_offering_count = 0)
            OR (NOT all_offerings AND expected_offering_count BETWEEN 1 AND 256))
        AND (valid_end IS NULL OR valid_start < valid_end)
    )
);

CREATE INDEX billing_discount_versions_validity_idx
    ON platform.billing_discount_versions (valid_start, valid_end, priority, id);
