-- +goose Up
CREATE TABLE iam_profile_backfill_plans (
    id uuid NOT NULL,
    resource_kind text NOT NULL,
    source_schema_id uuid NOT NULL,
    target_schema_id uuid NOT NULL,
    converter_version text COLLATE "C" NOT NULL,
    plan_sha256 bytea NOT NULL,
    source_count bigint NOT NULL,
    expected_source_sha256 bytea NOT NULL,
    expected_target_sha256 bytea NOT NULL,
    page_size integer NOT NULL,
    page_count integer NOT NULL,
    state text NOT NULL,
    activated_at timestamptz,
    fenced_at timestamptz,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_profile_backfill_plans_pkey PRIMARY KEY (id),
    CONSTRAINT iam_profile_backfill_plans_source_fk FOREIGN KEY (
        resource_kind, source_schema_id
    ) REFERENCES iam_profile_schema_versions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_plans_target_fk FOREIGN KEY (
        resource_kind, target_schema_id
    ) REFERENCES iam_profile_schema_versions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_plans_resource_check
        CHECK (resource_kind IN ('user', 'organization', 'role')),
    CONSTRAINT iam_profile_backfill_plans_identity_check
        CHECK (source_schema_id <> target_schema_id),
    CONSTRAINT iam_profile_backfill_plans_converter_check
        CHECK (converter_version = 'typed_map_v1'),
    CONSTRAINT iam_profile_backfill_plans_digest_check CHECK (
        pg_catalog.octet_length(plan_sha256) = 32 AND
        pg_catalog.octet_length(expected_source_sha256) = 32 AND
        pg_catalog.octet_length(expected_target_sha256) = 32
    ),
    CONSTRAINT iam_profile_backfill_plans_page_check CHECK (
        source_count >= 0 AND page_size BETWEEN 1 AND 250 AND
        page_count >= 1 AND revision > 0
    ),
    CONSTRAINT iam_profile_backfill_plans_state_check
        CHECK (state IN ('planned', 'running', 'verified', 'activated', 'fenced')),
    CONSTRAINT iam_profile_backfill_plans_lifecycle_check CHECK (
        (state IN ('planned', 'running', 'verified') AND
            activated_at IS NULL AND fenced_at IS NULL) OR
        (state = 'activated' AND activated_at IS NOT NULL AND fenced_at IS NULL) OR
        (state = 'fenced' AND activated_at IS NULL AND fenced_at IS NOT NULL)
    ),
    CONSTRAINT iam_profile_backfill_plans_time_check CHECK (
        updated_at >= created_at AND
        (activated_at IS NULL OR activated_at >= created_at) AND
        (fenced_at IS NULL OR fenced_at >= created_at)
    )
);
