-- +goose Up
CREATE TABLE platform.billing_price_versions (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    offering_version_id uuid NOT NULL,
    cadence text NOT NULL,
    minor_units bigint NOT NULL,
    currency character(3) COLLATE "C" NOT NULL,
    currency_registry_version bigint NOT NULL,
    currency_eligible boolean NOT NULL,
    effective_start timestamptz NOT NULL,
    effective_end timestamptz,
    money_policy_version text NOT NULL,
    published_at timestamptz NOT NULL,
    CONSTRAINT billing_price_versions_pkey PRIMARY KEY (id),
    CONSTRAINT billing_price_versions_natural_key UNIQUE (
        offering_version_id, cadence, currency, effective_start
    ),
    CONSTRAINT billing_price_versions_authority_id_key UNIQUE (authority_id, id),
    CONSTRAINT billing_price_versions_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_price_versions_offering_fk FOREIGN KEY (
        authority_id, offering_version_id
    ) REFERENCES platform.billing_offering_versions (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_price_versions_currency_fk FOREIGN KEY (
        currency_registry_version, currency, currency_eligible
    ) REFERENCES platform.billing_currency_exponents (
        registry_version, code, financial_money_eligible
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_price_versions_shape_check CHECK (
        id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND cadence IN ('one_time', 'monthly', 'yearly')
        AND minor_units BETWEEN 0 AND 9000000000000000
        AND currency_eligible
        AND currency_registry_version > 0
        AND money_policy_version = 'fc_money_v1'
    ),
    CONSTRAINT billing_price_versions_interval_check CHECK (
        effective_end IS NULL OR effective_start < effective_end
    )
);

CREATE TABLE platform.billing_entitlement_definitions (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    registry_version bigint NOT NULL,
    entitlement_key text COLLATE "C" NOT NULL,
    value_type text NOT NULL,
    meter_kind text,
    reset_policy text,
    enforcement text,
    unlimited_allowed boolean NOT NULL,
    owner_module text COLLATE "C" NOT NULL,
    unit text COLLATE "C" NOT NULL,
    source_contract text COLLATE "C" NOT NULL,
    canonical_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_entitlement_definitions_pkey PRIMARY KEY (id),
    CONSTRAINT billing_entitlement_definitions_key_key UNIQUE (entitlement_key),
    CONSTRAINT billing_entitlement_definitions_id_key_key
        UNIQUE (id, entitlement_key),
    CONSTRAINT billing_entitlement_definitions_type_key
        UNIQUE (id, value_type, unlimited_allowed),
    CONSTRAINT billing_entitlement_definitions_authority_id_key
        UNIQUE (authority_id, id),
    CONSTRAINT billing_entitlement_definitions_authority_type_key
        UNIQUE (authority_id, id, value_type, unlimited_allowed),
    CONSTRAINT billing_entitlement_definitions_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_entitlement_definitions_key_check CHECK (
        entitlement_key ~ '^[a-z](?:[a-z0-9_.]{0,126}[a-z0-9])?$'
        AND owner_module ~ '^[a-z][a-z0-9_]{0,62}[a-z0-9]$'
        AND source_contract = 'ADR-0082'
        AND registry_version > 0
        AND pg_catalog.octet_length(canonical_sha256) = 32
    ),
    CONSTRAINT billing_entitlement_definitions_type_check CHECK (
        (value_type = 'boolean' AND meter_kind IS NULL
            AND reset_policy IS NULL AND enforcement IS NULL
            AND NOT unlimited_allowed AND unit = 'enabled')
        OR (value_type = 'integer'
            AND meter_kind IN ('counter', 'gauge')
            AND reset_policy IN ('billing_period', 'calendar_month_utc', 'never')
            AND enforcement IN ('hard', 'soft')
            AND unit IN ('byte', 'count')
            AND (meter_kind <> 'gauge' OR reset_policy = 'never'))
    )
);
