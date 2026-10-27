-- +goose Up
CREATE TABLE platform.billing_revenue_projection (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    subscription_revision bigint NOT NULL,
    offering_version_id uuid NOT NULL,
    price_version_id uuid NOT NULL,
    price_source_sha256 bytea NOT NULL,
    state text COLLATE "C" NOT NULL,
    cadence text COLLATE "C" NOT NULL,
    currency character(3) COLLATE "C" NOT NULL,
    currency_registry_version bigint NOT NULL,
    source_minor_units bigint NOT NULL,
    monthly_minor_units bigint NOT NULL,
    annual_minor_units bigint NOT NULL,
    refreshed_at timestamptz NOT NULL,
    CONSTRAINT billing_revenue_projection_pkey PRIMARY KEY (id),
    CONSTRAINT billing_revenue_projection_subscription_key
        UNIQUE (subscription_id),
    CONSTRAINT billing_revenue_projection_authority_fk
        FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_revenue_projection_shape_check CHECK (
        id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND subscription_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND subscription_revision > 0
        AND pg_catalog.octet_length(price_source_sha256) = 32
        AND state IN ('active', 'past_due')
        AND cadence IN ('monthly', 'yearly')
        AND currency ~ '^[A-Z]{3}$'
        AND currency_registry_version > 0
        AND source_minor_units BETWEEN 0 AND 9000000000000000
        AND monthly_minor_units >= 0
        AND annual_minor_units >= 0
    )
);

CREATE INDEX billing_revenue_projection_group_idx
    ON platform.billing_revenue_projection (
        currency, currency_registry_version, state, cadence
    );

CREATE TABLE platform.billing_revenue_rebuilds (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    authority_group_generation bigint NOT NULL,
    billing_generation bigint NOT NULL,
    source_count bigint NOT NULL,
    projected_count bigint NOT NULL,
    source_revision_ceiling bigint NOT NULL,
    before_sha256 bytea NOT NULL,
    source_sha256 bytea NOT NULL,
    after_sha256 bytea NOT NULL,
    rebuilt_at timestamptz NOT NULL,
    CONSTRAINT billing_revenue_rebuilds_pkey PRIMARY KEY (id),
    CONSTRAINT billing_revenue_rebuilds_authority_fk
        FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_revenue_rebuilds_shape_check CHECK (
        id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND actor_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND authority_group_generation > 0 AND billing_generation > 0
        AND source_count >= 0 AND projected_count = source_count
        AND source_revision_ceiling >= 0
        AND pg_catalog.octet_length(before_sha256) = 32
        AND pg_catalog.octet_length(source_sha256) = 32
        AND pg_catalog.octet_length(after_sha256) = 32
        AND source_sha256 = after_sha256
    )
);
