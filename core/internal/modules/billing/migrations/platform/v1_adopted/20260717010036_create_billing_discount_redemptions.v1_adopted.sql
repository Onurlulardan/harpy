-- +goose Up
CREATE TABLE platform.billing_discount_offerings (
    authority_id uuid NOT NULL,
    discount_version_id uuid NOT NULL,
    offering_version_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_discount_offerings_pkey
        PRIMARY KEY (discount_version_id, offering_version_id),
    CONSTRAINT billing_discount_offerings_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_offerings_discount_fk FOREIGN KEY (
        authority_id, discount_version_id
    ) REFERENCES platform.billing_discount_versions (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_offerings_offering_fk FOREIGN KEY (
        authority_id, offering_version_id
    ) REFERENCES platform.billing_offering_versions (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE platform.billing_discount_redemptions (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    discount_version_id uuid NOT NULL,
    offering_version_id uuid NOT NULL,
    price_version_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    applied_kind text NOT NULL,
    applied_fixed_minor_units bigint,
    applied_currency character(3) COLLATE "C",
    applied_currency_registry_version bigint,
    applied_basis_points integer,
    applied_priority integer NOT NULL,
    applied_duration text NOT NULL,
    applied_repeating_cycles bigint,
    policy_sha256 bytea NOT NULL,
    redeemed_at timestamptz NOT NULL,
    relationship_ended_at timestamptz,
    CONSTRAINT billing_discount_redemptions_pkey PRIMARY KEY (id),
    CONSTRAINT billing_discount_redemptions_subscription_discount_key
        UNIQUE (subscription_id, discount_version_id),
    CONSTRAINT billing_discount_redemptions_operation_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_discount_redemptions_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_redemptions_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_redemptions_discount_fk
        FOREIGN KEY (discount_version_id)
        REFERENCES platform.billing_discount_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_redemptions_offering_fk
        FOREIGN KEY (offering_version_id)
        REFERENCES platform.billing_offering_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_redemptions_price_fk
        FOREIGN KEY (price_version_id)
        REFERENCES platform.billing_price_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_discount_redemptions_amount_check CHECK (
        (applied_kind = 'fixed' AND applied_fixed_minor_units > 0
            AND applied_currency IS NOT NULL
            AND applied_currency_registry_version > 0
            AND applied_basis_points IS NULL)
        OR (applied_kind = 'basis_point'
            AND applied_basis_points BETWEEN 1 AND 10000
            AND applied_fixed_minor_units IS NULL AND applied_currency IS NULL
            AND applied_currency_registry_version IS NULL)
    ),
    CONSTRAINT billing_discount_redemptions_duration_check CHECK (
        (applied_duration = 'once' AND applied_repeating_cycles IS NULL)
        OR (applied_duration = 'repeating' AND applied_repeating_cycles > 0)
        OR (applied_duration = 'forever' AND applied_repeating_cycles IS NULL)
    ),
    CONSTRAINT billing_discount_redemptions_time_check CHECK (
        applied_priority >= 0
        AND pg_catalog.octet_length(policy_sha256) = 32
        AND (relationship_ended_at IS NULL OR relationship_ended_at >= redeemed_at)
    )
);

CREATE INDEX billing_discount_redemptions_limit_idx
    ON platform.billing_discount_redemptions (
        discount_version_id, tenant_id, redeemed_at
    );
