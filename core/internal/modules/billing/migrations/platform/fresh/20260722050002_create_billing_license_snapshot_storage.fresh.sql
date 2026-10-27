-- +goose Up
CREATE TABLE platform.billing_license_snapshots (
    snapshot_id uuid NOT NULL DEFAULT pg_catalog.gen_random_uuid(),
    license_id uuid NOT NULL,
    licensee_organization_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    delivery_kind text NOT NULL,
    issuer text COLLATE "C" NOT NULL,
    key_id text COLLATE "C" NOT NULL,
    sequence bigint NOT NULL,
    payload_sha256 text COLLATE "C" NOT NULL,
    status text NOT NULL,
    issued_at timestamptz NOT NULL,
    not_before timestamptz NOT NULL,
    refresh_after timestamptz NOT NULL,
    grant_valid_until timestamptz NOT NULL,
    grace_until timestamptz NOT NULL,
    commercial_expires_at timestamptz NOT NULL,
    release_channel text COLLATE "C" NOT NULL,
    features jsonb NOT NULL,
    limits jsonb NOT NULL,
    canonical_payload bytea NOT NULL,
    signed_document bytea NOT NULL,
    accepted_at timestamptz NOT NULL,
    PRIMARY KEY (snapshot_id),
    CONSTRAINT billing_license_snapshot_sequence_uq
        UNIQUE (license_id, environment_id, sequence),
    CONSTRAINT billing_license_snapshot_digest_uq
        UNIQUE (license_id, environment_id, payload_sha256),
    CONSTRAINT billing_license_snapshot_head_fk_uq
        UNIQUE (snapshot_id, license_id, environment_id, sequence, payload_sha256),
    CONSTRAINT billing_license_snapshot_environment_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.environment_identity (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_snapshot_shape_check CHECK (
        license_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND licensee_organization_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND delivery_kind IN ('connected', 'airgap', 'development')
        AND pg_catalog.octet_length(issuer) BETWEEN 1 AND 256
        AND issuer = pg_catalog.btrim(issuer)
        AND issuer !~ '[[:cntrl:]]'
        AND key_id ~ '^[a-z0-9][a-z0-9._-]{0,63}$'
        AND sequence > 0 AND payload_sha256 ~ '^[0-9a-f]{64}$'
        AND status IN ('active', 'suspended', 'revoked', 'expired')
        AND release_channel ~ '^[a-z0-9][a-z0-9._-]{0,63}$'
        AND pg_catalog.jsonb_typeof(features) = 'object'
        AND pg_catalog.jsonb_typeof(limits) = 'object'
        AND pg_catalog.octet_length(canonical_payload) BETWEEN 2 AND 65536
        AND pg_catalog.octet_length(signed_document) BETWEEN 2 AND 65536
        AND payload_sha256 = pg_catalog.encode(
            pg_catalog.sha256(canonical_payload), 'hex'
        )
    ),
    CONSTRAINT billing_license_snapshot_time_check CHECK (
        issued_at <= refresh_after AND not_before <= grant_valid_until
        AND grant_valid_until <= grace_until
        AND not_before <= commercial_expires_at
    )
);

CREATE TABLE platform.billing_license_snapshot_projections (
    snapshot_id uuid NOT NULL,
    entitlement_key text COLLATE "C" NOT NULL,
    value_type text NOT NULL,
    boolean_value boolean,
    integer_value bigint,
    unlimited boolean NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (snapshot_id, entitlement_key),
    CONSTRAINT billing_license_projection_snapshot_fk
        FOREIGN KEY (snapshot_id)
        REFERENCES platform.billing_license_snapshots (snapshot_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_projection_key_check CHECK (
        entitlement_key ~ '^[a-z](?:[a-z0-9_.]{0,126}[a-z0-9])?$'
    ),
    CONSTRAINT billing_license_projection_value_check CHECK (
        (value_type = 'boolean' AND boolean_value IS NOT NULL
            AND integer_value IS NULL AND NOT unlimited)
        OR (value_type = 'integer' AND boolean_value IS NULL
            AND ((unlimited AND integer_value IS NULL)
                OR (NOT unlimited AND integer_value >= 0)))
    )
);

CREATE INDEX billing_license_snapshot_projection_key_idx
ON platform.billing_license_snapshot_projections (
    entitlement_key, snapshot_id
);
