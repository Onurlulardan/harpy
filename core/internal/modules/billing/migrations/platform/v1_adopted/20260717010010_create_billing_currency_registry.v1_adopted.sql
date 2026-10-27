-- +goose Up
CREATE TABLE platform.billing_currency_registry_versions (
    authority_id uuid NOT NULL,
    registry_version bigint NOT NULL,
    source_published_on date NOT NULL,
    source_url text NOT NULL,
    source_bytes bigint NOT NULL,
    source_sha256 text NOT NULL,
    normalized_sha256 text NOT NULL,
    unique_code_count integer NOT NULL,
    not_applicable_code_count integer NOT NULL,
    published_at timestamptz NOT NULL,
    CONSTRAINT billing_currency_registry_versions_pkey
        PRIMARY KEY (registry_version),
    CONSTRAINT billing_currency_registry_versions_authority_key
        UNIQUE (authority_id, registry_version),
    CONSTRAINT billing_currency_registry_versions_authority_fk
        FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_currency_registry_versions_version_check
        CHECK (registry_version > 0),
    CONSTRAINT billing_currency_registry_versions_source_check CHECK (
        source_bytes > 0
        AND pg_catalog.octet_length(source_url) BETWEEN 1 AND 2048
        AND source_sha256 ~ '^[0-9a-f]{64}$'
        AND normalized_sha256 ~ '^[0-9a-f]{64}$'
        AND unique_code_count > 0
        AND not_applicable_code_count BETWEEN 0 AND unique_code_count
    )
);

CREATE TABLE platform.billing_currency_exponents (
    authority_id uuid NOT NULL,
    registry_version bigint NOT NULL,
    code character(3) COLLATE "C" NOT NULL,
    numeric_code character(3) COLLATE "C",
    exponent smallint,
    financial_money_eligible boolean NOT NULL,
    CONSTRAINT billing_currency_exponents_pkey
        PRIMARY KEY (registry_version, code),
    CONSTRAINT billing_currency_exponents_eligibility_key
        UNIQUE (registry_version, code, financial_money_eligible),
    CONSTRAINT billing_currency_exponents_registry_fk FOREIGN KEY (
        authority_id, registry_version
    ) REFERENCES platform.billing_currency_registry_versions (
        authority_id, registry_version
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_currency_exponents_code_check CHECK (
        code ~ '^[A-Z]{3}$'
        AND (numeric_code IS NULL OR numeric_code ~ '^[0-9]{3}$')
    ),
    CONSTRAINT billing_currency_exponents_value_check CHECK (
        (financial_money_eligible AND exponent IN (0, 2, 3, 4))
        OR (NOT financial_money_eligible AND exponent IS NULL)
    )
);

