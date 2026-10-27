-- +goose Up
CREATE TABLE iam_maintenance_audit (
    job_id uuid NOT NULL,
    attempt_id uuid NOT NULL,
    kind text NOT NULL,
    outcome_code text COLLATE "C" NOT NULL,
    affected_count bigint NOT NULL,
    error_count bigint NOT NULL,
    result_sha256 bytea NOT NULL,
    occurred_at timestamptz NOT NULL,
    CONSTRAINT iam_maintenance_audit_pkey PRIMARY KEY (job_id),
    CONSTRAINT iam_maintenance_audit_attempt_key UNIQUE (attempt_id),
    CONSTRAINT iam_maintenance_audit_receipt_fk FOREIGN KEY (job_id)
        REFERENCES iam_maintenance_receipts (job_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_maintenance_audit_redacted_check CHECK (
        affected_count >= 0 AND error_count >= 0 AND
        pg_catalog.octet_length(result_sha256) = 32 AND
        pg_catalog.octet_length(outcome_code) BETWEEN 1 AND 128
    )
);

CREATE TABLE iam_maintenance_metrics (
    kind text NOT NULL,
    completed_runs bigint NOT NULL,
    affected_total bigint NOT NULL,
    error_total bigint NOT NULL,
    last_job_id uuid NOT NULL,
    last_result_sha256 bytea NOT NULL,
    last_completed_at timestamptz NOT NULL,
    revision bigint NOT NULL,
    CONSTRAINT iam_maintenance_metrics_pkey PRIMARY KEY (kind),
    CONSTRAINT iam_maintenance_metrics_count_check CHECK (
        completed_runs > 0 AND affected_total >= 0 AND error_total >= 0 AND
        revision > 0
    ),
    CONSTRAINT iam_maintenance_metrics_digest_check
        CHECK (pg_catalog.octet_length(last_result_sha256) = 32)
);
