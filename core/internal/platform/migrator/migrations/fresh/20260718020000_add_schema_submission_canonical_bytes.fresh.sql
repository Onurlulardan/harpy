-- +goose Up
ALTER TABLE platform.schema_jobs
ADD COLUMN canonical_plan_bytes bytea;

ALTER TABLE platform.schema_jobs
ADD CONSTRAINT schema_jobs_canonical_bytes_check CHECK (
    canonical_plan_bytes IS NULL OR (
        pg_catalog.octet_length(canonical_plan_bytes) BETWEEN 2 AND 1048576
        AND pg_catalog.pg_input_is_valid(
            pg_catalog.convert_from(canonical_plan_bytes, 'UTF8'), 'jsonb'
        )
        AND pg_catalog.convert_from(canonical_plan_bytes, 'UTF8')::jsonb
            = canonical_plan
        AND pg_catalog.encode(
            pg_catalog.sha256(canonical_plan_bytes), 'hex'
        ) = plan_sha256
    )
);

ALTER TABLE platform.schema_job_steps
ADD COLUMN canonical_step_bytes bytea;

ALTER TABLE platform.schema_job_steps
ADD CONSTRAINT schema_job_steps_canonical_bytes_check CHECK (
    canonical_step_bytes IS NULL OR (
        pg_catalog.octet_length(canonical_step_bytes) BETWEEN 2 AND 1048576
        AND pg_catalog.pg_input_is_valid(
            pg_catalog.convert_from(canonical_step_bytes, 'UTF8'), 'jsonb'
        )
        AND pg_catalog.encode(
            pg_catalog.sha256(canonical_step_bytes), 'hex'
        ) = step_sha256
    )
);
