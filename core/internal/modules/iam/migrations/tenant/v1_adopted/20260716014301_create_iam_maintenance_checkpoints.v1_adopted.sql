-- +goose Up
CREATE TABLE iam_maintenance_checkpoints (
    kind text NOT NULL,
    cursor_id uuid,
    scanned_count bigint NOT NULL,
    affected_count bigint NOT NULL,
    error_count bigint NOT NULL,
    result_sha256 bytea NOT NULL,
    last_job_id uuid,
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_maintenance_checkpoints_pkey PRIMARY KEY (kind),
    CONSTRAINT iam_maintenance_checkpoints_schedule_fk FOREIGN KEY (kind)
        REFERENCES iam_maintenance_schedules (kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_maintenance_checkpoints_count_check CHECK (
        scanned_count >= 0 AND affected_count >= 0 AND error_count >= 0 AND
        affected_count <= scanned_count AND revision > 0
    ),
    CONSTRAINT iam_maintenance_checkpoints_digest_check
        CHECK (pg_catalog.octet_length(result_sha256) = 32)
);
