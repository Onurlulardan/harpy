-- +goose Up
CREATE TABLE iam_profile_backfill_mappings (
    plan_id uuid NOT NULL,
    ordinal integer NOT NULL,
    source_field_id uuid NOT NULL,
    target_field_id uuid NOT NULL,
    converter_kind text COLLATE "C" NOT NULL,
    semantic_sha256 bytea NOT NULL,
    CONSTRAINT iam_profile_backfill_mappings_pkey PRIMARY KEY (plan_id, ordinal),
    CONSTRAINT iam_profile_backfill_mappings_source_key
        UNIQUE (plan_id, source_field_id),
    CONSTRAINT iam_profile_backfill_mappings_target_key
        UNIQUE (plan_id, target_field_id),
    CONSTRAINT iam_profile_backfill_mappings_plan_fk FOREIGN KEY (plan_id)
        REFERENCES iam_profile_backfill_plans (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_mappings_ordinal_check
        CHECK (ordinal BETWEEN 1 AND 512),
    CONSTRAINT iam_profile_backfill_mappings_identity_check
        CHECK (source_field_id <> target_field_id OR converter_kind = 'copy_v1'),
    CONSTRAINT iam_profile_backfill_mappings_converter_check
        CHECK (converter_kind IN ('copy_v1', 'copy_normalize_v1')),
    CONSTRAINT iam_profile_backfill_mappings_digest_check
        CHECK (pg_catalog.octet_length(semantic_sha256) = 32)
);
