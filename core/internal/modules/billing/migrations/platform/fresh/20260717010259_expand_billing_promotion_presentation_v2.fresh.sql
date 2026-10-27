-- +goose Up
ALTER TABLE platform.billing_discount_versions
    ADD COLUMN presentation_name text,
    ADD COLUMN presentation_description text,
    ADD COLUMN available boolean;

UPDATE platform.billing_discount_versions AS version_row
SET presentation_name = root.normalized_code,
    presentation_description = '', available = (root.lifecycle = 'active')
FROM platform.billing_discounts AS root
WHERE root.authority_id = version_row.authority_id
  AND root.id = version_row.discount_id;

ALTER TABLE platform.billing_discount_versions
    ALTER COLUMN presentation_name SET NOT NULL,
    ALTER COLUMN presentation_description SET NOT NULL,
    ALTER COLUMN available SET NOT NULL,
    ADD CONSTRAINT billing_discount_versions_presentation_v2_check CHECK (
        pg_catalog.octet_length(presentation_name) BETWEEN 1 AND 100
        AND pg_catalog.btrim(presentation_name) = presentation_name
        AND pg_catalog.octet_length(presentation_description) <= 4096
    );

ALTER TABLE platform.billing_campaign_versions
    ADD COLUMN presentation_name text,
    ADD COLUMN presentation_description text,
    ADD COLUMN presentation_banner_text text,
    ADD COLUMN auto_apply boolean,
    ADD COLUMN available boolean,
    ADD COLUMN superseded_version_id uuid;

UPDATE platform.billing_campaign_versions AS version_row
SET presentation_name = version_row.presentation_metadata ->> 'name',
    presentation_description = COALESCE(version_row.presentation_metadata ->> 'description', ''),
    presentation_banner_text = version_row.presentation_metadata ->> 'banner_text',
    auto_apply = COALESCE((version_row.presentation_metadata ->> 'auto_apply')::boolean, false),
    available = (root.lifecycle = 'active')
FROM platform.billing_campaigns AS root
WHERE root.authority_id = version_row.authority_id
  AND root.id = version_row.campaign_id;

UPDATE platform.billing_campaign_versions
SET presentation_metadata = pg_catalog.jsonb_build_object(
        'name', presentation_name,
        'description', presentation_description,
        'banner_text', presentation_banner_text,
        'auto_apply', auto_apply
    );
UPDATE platform.billing_campaign_versions
SET presentation_sha256 = pg_catalog.sha256(
    pg_catalog.convert_to(presentation_metadata::text, 'UTF8')
);

ALTER TABLE platform.billing_campaign_versions
    ALTER COLUMN presentation_name SET NOT NULL,
    ALTER COLUMN presentation_description SET NOT NULL,
    ALTER COLUMN presentation_banner_text SET NOT NULL,
    ALTER COLUMN auto_apply SET NOT NULL,
    ALTER COLUMN available SET NOT NULL,
    DROP CONSTRAINT billing_campaign_versions_presentation_check,
    ADD CONSTRAINT billing_campaign_versions_presentation_v2_check CHECK (
        pg_catalog.octet_length(presentation_name) BETWEEN 1 AND 128
        AND pg_catalog.btrim(presentation_name) = presentation_name
        AND pg_catalog.octet_length(presentation_description) <= 4096
        AND pg_catalog.octet_length(presentation_banner_text) <= 512
        AND presentation_metadata = pg_catalog.jsonb_build_object(
            'name', presentation_name,
            'description', presentation_description,
            'banner_text', presentation_banner_text,
            'auto_apply', auto_apply
        )
        AND pg_catalog.octet_length(presentation_sha256) = 32
        AND presentation_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(presentation_metadata::text, 'UTF8')
        )
    ),
    ADD CONSTRAINT billing_campaign_versions_superseded_fk FOREIGN KEY (
        authority_id, superseded_version_id
    ) REFERENCES platform.billing_campaign_versions (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    ADD CONSTRAINT billing_campaign_versions_superseded_key
        UNIQUE (superseded_version_id);

CREATE INDEX billing_campaign_versions_current_v2_idx
    ON platform.billing_campaign_versions (campaign_id, version DESC, id)
    WHERE superseded_version_id IS NULL;

CREATE TABLE platform.billing_promotion_association_receipts (
    operation_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    discount_id uuid NOT NULL,
    previous_campaign_id uuid,
    campaign_id uuid,
    occurred_at timestamptz NOT NULL,
    CONSTRAINT billing_promotion_association_receipts_pkey PRIMARY KEY (operation_id),
    CONSTRAINT billing_promotion_association_receipts_operation_fk FOREIGN KEY (
        operation_id
    ) REFERENCES platform.billing_catalog_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_promotion_association_receipts_authority_fk FOREIGN KEY (
        authority_id
    ) REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_promotion_association_receipts_discount_fk FOREIGN KEY (
        authority_id, discount_id
    ) REFERENCES platform.billing_discounts (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_promotion_association_receipts_previous_fk FOREIGN KEY (
        authority_id, previous_campaign_id
    ) REFERENCES platform.billing_campaigns (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_promotion_association_receipts_campaign_fk FOREIGN KEY (
        authority_id, campaign_id
    ) REFERENCES platform.billing_campaigns (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_promotion_association_receipts_change_check CHECK (
        previous_campaign_id IS DISTINCT FROM campaign_id
    )
);
