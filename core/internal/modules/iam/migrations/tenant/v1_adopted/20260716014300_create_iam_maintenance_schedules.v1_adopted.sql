-- +goose Up
CREATE TABLE iam_maintenance_schedules (
    kind text NOT NULL,
    payload_revision integer NOT NULL,
    payload_reference text COLLATE "C" NOT NULL,
    payload_sha256 bytea NOT NULL,
    current_job_id uuid NOT NULL,
    scheduled_run_at timestamptz NOT NULL,
    cadence_seconds integer NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_maintenance_schedules_pkey PRIMARY KEY (kind),
    CONSTRAINT iam_maintenance_schedules_job_key UNIQUE (current_job_id),
    CONSTRAINT iam_maintenance_schedules_kind_check CHECK (kind IN (
        'iam.directory.repair', 'iam.directory.checksum',
        'iam.credential.bcrypt-retire', 'iam.identifier.retire',
        'iam.directory.tombstone-purge', 'iam.security-artifact-purge'
    )),
    CONSTRAINT iam_maintenance_schedules_revision_check
        CHECK (payload_revision = 1 AND revision > 0),
    CONSTRAINT iam_maintenance_schedules_reference_check CHECK (
        pg_catalog.octet_length(payload_reference) BETWEEN 1 AND 1024 AND
        pg_catalog.octet_length(payload_sha256) = 32 AND
        pg_catalog.sha256(pg_catalog.convert_to(payload_reference, 'UTF8')) =
            payload_sha256
    ),
    CONSTRAINT iam_maintenance_schedules_cadence_check CHECK (
        (kind = 'iam.directory.repair' AND cadence_seconds = 900) OR
        (kind = 'iam.security-artifact-purge' AND cadence_seconds = 3600) OR
        (kind IN (
            'iam.directory.checksum', 'iam.credential.bcrypt-retire',
            'iam.identifier.retire', 'iam.directory.tombstone-purge'
        ) AND cadence_seconds = 86400)
    ),
    CONSTRAINT iam_maintenance_schedules_time_check
        CHECK (updated_at >= created_at)
);

CREATE INDEX iam_maintenance_schedules_due_idx
    ON iam_maintenance_schedules (scheduled_run_at, kind);
