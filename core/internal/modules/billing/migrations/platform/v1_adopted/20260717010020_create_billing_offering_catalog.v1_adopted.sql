-- +goose Up
CREATE TABLE platform.billing_offerings (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    offering_key text COLLATE "C" NOT NULL,
    lifecycle text NOT NULL,
    created_at timestamptz NOT NULL,
    retired_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_offerings_pkey PRIMARY KEY (id),
    CONSTRAINT billing_offerings_key_key UNIQUE (offering_key),
    CONSTRAINT billing_offerings_authority_id_key UNIQUE (authority_id, id),
    CONSTRAINT billing_offerings_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_offerings_id_check CHECK (
        id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT billing_offerings_key_check CHECK (
        offering_key ~ '^[a-z][a-z0-9._-]{0,126}[a-z0-9]$'
    ),
    CONSTRAINT billing_offerings_lifecycle_check CHECK (
        (lifecycle = 'active' AND retired_at IS NULL)
        OR (lifecycle = 'retired' AND retired_at IS NOT NULL)
    ),
    CONSTRAINT billing_offerings_revision_check CHECK (revision > 0),
    CONSTRAINT billing_offerings_time_check CHECK (
        retired_at IS NULL OR retired_at >= created_at
    )
);

CREATE TABLE platform.billing_offering_versions (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    offering_id uuid NOT NULL,
    version bigint NOT NULL,
    is_base boolean NOT NULL,
    is_public boolean NOT NULL,
    trial_days integer NOT NULL,
    grace_days integer NOT NULL,
    expected_price_count smallint NOT NULL,
    expected_grant_count smallint NOT NULL,
    effective_start timestamptz NOT NULL,
    effective_end timestamptz,
    presentation_metadata jsonb NOT NULL,
    presentation_sha256 bytea NOT NULL,
    publication_state text NOT NULL,
    publication_operation_id uuid NOT NULL,
    published_at timestamptz NOT NULL,
    CONSTRAINT billing_offering_versions_pkey PRIMARY KEY (id),
    CONSTRAINT billing_offering_versions_offering_version_key
        UNIQUE (offering_id, version),
    CONSTRAINT billing_offering_versions_operation_key
        UNIQUE (publication_operation_id),
    CONSTRAINT billing_offering_versions_authority_id_key UNIQUE (authority_id, id),
    CONSTRAINT billing_offering_versions_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_offering_versions_offering_fk FOREIGN KEY (
        authority_id, offering_id
    ) REFERENCES platform.billing_offerings (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_offering_versions_id_check CHECK (
        id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT billing_offering_versions_policy_check CHECK (
        publication_operation_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND version > 0 AND trial_days BETWEEN 0 AND 3650
        AND grace_days BETWEEN 0 AND 3650
        AND expected_price_count BETWEEN 1 AND 16
        AND expected_grant_count BETWEEN 0 AND 256
        AND publication_state = 'published'
    ),
    CONSTRAINT billing_offering_versions_interval_check CHECK (
        effective_end IS NULL OR effective_start < effective_end
    ),
    CONSTRAINT billing_offering_versions_presentation_check CHECK (
        pg_catalog.jsonb_typeof(presentation_metadata) = 'object'
        AND pg_catalog.octet_length(presentation_sha256) = 32
        AND presentation_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(presentation_metadata::text, 'UTF8')
        )
    )
);

CREATE INDEX billing_offering_versions_public_idx
    ON platform.billing_offering_versions (
        is_public, effective_start, effective_end, offering_id, version
    ) WHERE is_public;
