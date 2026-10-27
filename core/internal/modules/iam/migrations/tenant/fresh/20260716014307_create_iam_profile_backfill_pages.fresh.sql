-- +goose Up
CREATE TABLE iam_profile_backfill_pages (
    plan_id uuid NOT NULL,
    page_number integer NOT NULL,
    job_id uuid NOT NULL,
    scheduled_run_at timestamptz NOT NULL,
    payload_reference text COLLATE "C" NOT NULL,
    payload_sha256 bytea NOT NULL,
    after_resource_id uuid,
    through_resource_id uuid,
    expected_count integer NOT NULL,
    source_sha256 bytea NOT NULL,
    expected_target_sha256 bytea NOT NULL,
    state text NOT NULL,
    converted_count integer NOT NULL,
    error_count integer NOT NULL,
    result_sha256 bytea,
    completed_at timestamptz,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_profile_backfill_pages_pkey PRIMARY KEY (plan_id, page_number),
    CONSTRAINT iam_profile_backfill_pages_job_key UNIQUE (job_id),
    CONSTRAINT iam_profile_backfill_pages_plan_fk FOREIGN KEY (plan_id)
        REFERENCES iam_profile_backfill_plans (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_pages_number_check
        CHECK (page_number > 0 AND expected_count BETWEEN 0 AND 250),
    CONSTRAINT iam_profile_backfill_pages_reference_check CHECK (
        pg_catalog.octet_length(payload_reference) BETWEEN 1 AND 1024 AND
        pg_catalog.octet_length(payload_sha256) = 32 AND
        pg_catalog.sha256(pg_catalog.convert_to(payload_reference, 'UTF8')) =
            payload_sha256
    ),
    CONSTRAINT iam_profile_backfill_pages_range_check CHECK (
        through_resource_id IS NOT NULL AND
        (after_resource_id IS NULL OR after_resource_id < through_resource_id)
    ),
    CONSTRAINT iam_profile_backfill_pages_digest_check CHECK (
        pg_catalog.octet_length(source_sha256) = 32 AND
        pg_catalog.octet_length(expected_target_sha256) = 32 AND
        (result_sha256 IS NULL OR pg_catalog.octet_length(result_sha256) = 32)
    ),
    CONSTRAINT iam_profile_backfill_pages_state_check
        CHECK (state IN ('queued', 'completed', 'fenced')),
    CONSTRAINT iam_profile_backfill_pages_result_check CHECK (
        converted_count >= 0 AND error_count >= 0 AND
        converted_count + error_count <= expected_count AND
        (state = 'queued' AND converted_count = 0 AND error_count = 0 AND
            result_sha256 IS NULL AND completed_at IS NULL OR
         state IN ('completed', 'fenced') AND result_sha256 IS NOT NULL AND
            completed_at IS NOT NULL)
    )
);

CREATE TABLE iam_profile_backfill_checkpoints (
    plan_id uuid NOT NULL,
    completed_pages integer NOT NULL,
    converted_count bigint NOT NULL,
    error_count bigint NOT NULL,
    rolling_sha256 bytea NOT NULL,
    last_page_number integer,
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_profile_backfill_checkpoints_pkey PRIMARY KEY (plan_id),
    CONSTRAINT iam_profile_backfill_checkpoints_plan_fk FOREIGN KEY (plan_id)
        REFERENCES iam_profile_backfill_plans (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_checkpoints_count_check CHECK (
        completed_pages >= 0 AND converted_count >= 0 AND error_count >= 0 AND
        revision > 0 AND
        ((completed_pages = 0 AND last_page_number IS NULL) OR
         (completed_pages > 0 AND last_page_number = completed_pages))
    ),
    CONSTRAINT iam_profile_backfill_checkpoints_digest_check
        CHECK (pg_catalog.octet_length(rolling_sha256) = 32)
);
