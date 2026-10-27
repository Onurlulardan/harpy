-- +goose Up
CREATE TABLE platform.billing_campaigns (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    campaign_key text COLLATE "C" NOT NULL,
    lifecycle text NOT NULL,
    created_at timestamptz NOT NULL,
    retired_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_campaigns_pkey PRIMARY KEY (id),
    CONSTRAINT billing_campaigns_key_key UNIQUE (campaign_key),
    CONSTRAINT billing_campaigns_authority_id_key UNIQUE (authority_id, id),
    CONSTRAINT billing_campaigns_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_campaigns_key_check CHECK (
        campaign_key ~ '^[a-z][a-z0-9._-]{0,126}[a-z0-9]$'
    ),
    CONSTRAINT billing_campaigns_lifecycle_check CHECK (
        (lifecycle = 'active' AND retired_at IS NULL)
        OR (lifecycle = 'retired' AND retired_at IS NOT NULL)
    ),
    CONSTRAINT billing_campaigns_shape_check CHECK (
        id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND revision > 0
        AND (retired_at IS NULL OR retired_at >= created_at)
    )
);

CREATE TABLE platform.billing_campaign_versions (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    campaign_id uuid NOT NULL,
    version bigint NOT NULL,
    priority integer NOT NULL,
    expected_discount_count smallint NOT NULL,
    starts_at timestamptz NOT NULL,
    ends_at timestamptz,
    presentation_metadata jsonb NOT NULL,
    presentation_sha256 bytea NOT NULL,
    publication_state text NOT NULL,
    publication_operation_id uuid NOT NULL,
    published_at timestamptz NOT NULL,
    CONSTRAINT billing_campaign_versions_pkey PRIMARY KEY (id),
    CONSTRAINT billing_campaign_versions_campaign_version_key
        UNIQUE (campaign_id, version),
    CONSTRAINT billing_campaign_versions_authority_id_key UNIQUE (authority_id, id),
    CONSTRAINT billing_campaign_versions_operation_key
        UNIQUE (publication_operation_id),
    CONSTRAINT billing_campaign_versions_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_campaign_versions_campaign_fk FOREIGN KEY (
        authority_id, campaign_id
    ) REFERENCES platform.billing_campaigns (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_campaign_versions_policy_check CHECK (
        id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND publication_operation_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND version > 0 AND expected_discount_count BETWEEN 0 AND 256
        AND publication_state = 'published'
    ),
    CONSTRAINT billing_campaign_versions_interval_check CHECK (
        ends_at IS NULL OR starts_at < ends_at
    ),
    CONSTRAINT billing_campaign_versions_presentation_check CHECK (
        pg_catalog.jsonb_typeof(presentation_metadata) = 'object'
        AND pg_catalog.octet_length(presentation_sha256) = 32
        AND presentation_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(presentation_metadata::text, 'UTF8')
        )
    )
);

CREATE INDEX billing_campaign_versions_active_lookup_idx
    ON platform.billing_campaign_versions (priority DESC, id, starts_at, ends_at)
    WHERE publication_state = 'published';

CREATE TABLE platform.billing_campaign_discounts (
    authority_id uuid NOT NULL,
    campaign_version_id uuid NOT NULL,
    discount_version_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_campaign_discounts_pkey
        PRIMARY KEY (campaign_version_id, discount_version_id),
    CONSTRAINT billing_campaign_discounts_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_campaign_discounts_campaign_fk FOREIGN KEY (
        authority_id, campaign_version_id
    ) REFERENCES platform.billing_campaign_versions (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_campaign_discounts_discount_fk FOREIGN KEY (
        authority_id, discount_version_id
    ) REFERENCES platform.billing_discount_versions (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE INDEX billing_campaign_discounts_order_idx
    ON platform.billing_campaign_discounts (campaign_version_id, discount_version_id);
