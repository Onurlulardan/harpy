-- +goose Up
CREATE UNIQUE INDEX iam_profile_backfill_plans_one_live_resource_idx
    ON iam_profile_backfill_plans (resource_kind)
    WHERE state IN ('planned', 'running', 'verified');

CREATE TABLE iam_profile_backfill_preconversions (
    plan_id uuid NOT NULL,
    page_number integer NOT NULL,
    resource_id uuid NOT NULL,
    resource_kind text NOT NULL,
    actual_source_schema_id uuid NOT NULL,
    source_revision bigint NOT NULL,
    source_sha256 bytea NOT NULL,
    planned_target_sha256 bytea NOT NULL,
    first_target_revision bigint NOT NULL,
    first_target_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_profile_backfill_preconversions_pkey
        PRIMARY KEY (plan_id, resource_id),
    CONSTRAINT iam_profile_backfill_preconversions_page_fk
        FOREIGN KEY (plan_id, page_number)
        REFERENCES iam_profile_backfill_pages (plan_id, page_number)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_preconversions_schema_fk
        FOREIGN KEY (resource_kind, actual_source_schema_id)
        REFERENCES iam_profile_schema_versions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_preconversions_resource_check
        CHECK (resource_kind IN ('user', 'organization', 'role')),
    CONSTRAINT iam_profile_backfill_preconversions_revision_check CHECK (
        source_revision > 0 AND
        first_target_revision = source_revision + 1
    ),
    CONSTRAINT iam_profile_backfill_preconversions_digest_check CHECK (
        pg_catalog.octet_length(source_sha256) = 32 AND
        pg_catalog.octet_length(planned_target_sha256) = 32 AND
        pg_catalog.octet_length(first_target_sha256) = 32
    )
);
