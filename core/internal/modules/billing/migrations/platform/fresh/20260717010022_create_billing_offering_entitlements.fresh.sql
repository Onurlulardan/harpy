-- +goose Up
CREATE TABLE platform.billing_offering_entitlements (
    authority_id uuid NOT NULL,
    offering_version_id uuid NOT NULL,
    entitlement_definition_id uuid NOT NULL,
    value_type text NOT NULL,
    definition_allows_unlimited boolean NOT NULL,
    boolean_value boolean,
    integer_value bigint,
    unlimited boolean NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_offering_entitlements_pkey PRIMARY KEY (
        offering_version_id, entitlement_definition_id
    ),
    CONSTRAINT billing_offering_entitlements_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_offering_entitlements_offering_fk FOREIGN KEY (
        authority_id, offering_version_id
    ) REFERENCES platform.billing_offering_versions (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_offering_entitlements_definition_fk FOREIGN KEY (
        authority_id, entitlement_definition_id,
        value_type, definition_allows_unlimited
    ) REFERENCES platform.billing_entitlement_definitions (
        authority_id, id, value_type, unlimited_allowed
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_offering_entitlements_value_check CHECK (
        (value_type = 'boolean' AND boolean_value IS NOT NULL
            AND integer_value IS NULL AND NOT unlimited)
        OR (value_type = 'integer' AND boolean_value IS NULL AND (
            (unlimited AND definition_allows_unlimited AND integer_value IS NULL)
            OR (NOT unlimited AND integer_value >= 0)
        ))
    )
);

CREATE INDEX billing_offering_entitlements_definition_idx
    ON platform.billing_offering_entitlements (
        entitlement_definition_id, offering_version_id
    );
