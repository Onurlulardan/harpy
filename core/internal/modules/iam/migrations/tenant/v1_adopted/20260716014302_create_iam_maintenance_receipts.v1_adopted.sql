-- +goose Up
CREATE TABLE iam_maintenance_receipts (
    job_id uuid NOT NULL,
    attempt_id uuid NOT NULL,
    kind text NOT NULL,
    payload_revision integer NOT NULL,
    payload_reference text COLLATE "C" NOT NULL,
    payload_sha256 bytea NOT NULL,
    outcome_code text COLLATE "C" NOT NULL,
    scanned_count bigint NOT NULL,
    affected_count bigint NOT NULL,
    error_count bigint NOT NULL,
    result_sha256 bytea NOT NULL,
    successor_job_id uuid,
    successor_run_at timestamptz,
    successor_payload_reference text COLLATE "C",
    successor_payload_sha256 bytea,
    successor_dedupe_key text COLLATE "C",
    successor_priority integer,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_maintenance_receipts_pkey PRIMARY KEY (job_id),
    CONSTRAINT iam_maintenance_receipts_attempt_key UNIQUE (attempt_id),
    CONSTRAINT iam_maintenance_receipts_successor_key UNIQUE (successor_job_id),
    CONSTRAINT iam_maintenance_receipts_kind_check CHECK (kind IN (
        'iam.directory.repair', 'iam.directory.checksum',
        'iam.credential.bcrypt-retire', 'iam.identifier.retire',
        'iam.directory.tombstone-purge', 'iam.security-artifact-purge',
        'iam.profile.backfill'
    )),
    CONSTRAINT iam_maintenance_receipts_revision_check
        CHECK (payload_revision = 1),
    CONSTRAINT iam_maintenance_receipts_reference_check CHECK (
        pg_catalog.octet_length(payload_reference) BETWEEN 1 AND 1024 AND
        pg_catalog.octet_length(payload_sha256) = 32 AND
        pg_catalog.sha256(pg_catalog.convert_to(payload_reference, 'UTF8')) =
            payload_sha256
    ),
    CONSTRAINT iam_maintenance_receipts_outcome_check CHECK (
        outcome_code ~ '^iam[.][a-z0-9._-]+[.]completed$' AND
        pg_catalog.octet_length(outcome_code) BETWEEN 1 AND 128
    ),
    CONSTRAINT iam_maintenance_receipts_count_check CHECK (
        scanned_count >= 0 AND affected_count >= 0 AND error_count >= 0 AND
        affected_count <= scanned_count
    ),
    CONSTRAINT iam_maintenance_receipts_digest_check
        CHECK (pg_catalog.octet_length(result_sha256) = 32),
    CONSTRAINT iam_maintenance_receipts_successor_check CHECK (
        (successor_job_id IS NULL AND successor_run_at IS NULL AND
            successor_payload_reference IS NULL AND
            successor_payload_sha256 IS NULL AND successor_dedupe_key IS NULL AND
            successor_priority IS NULL) OR
        (successor_job_id IS NOT NULL AND successor_run_at IS NOT NULL AND
            pg_catalog.octet_length(successor_payload_reference) BETWEEN 1 AND 1024 AND
            pg_catalog.octet_length(successor_payload_sha256) = 32 AND
            pg_catalog.sha256(pg_catalog.convert_to(
                successor_payload_reference, 'UTF8'
            )) = successor_payload_sha256 AND
            pg_catalog.octet_length(successor_dedupe_key) BETWEEN 1 AND 256 AND
            successor_priority IS NOT NULL)
    )
);
