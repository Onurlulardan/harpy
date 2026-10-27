-- +goose Up
CREATE TABLE iam_profile_backfill_errors (
    plan_id uuid NOT NULL,
    page_number integer NOT NULL,
    resource_id uuid NOT NULL,
    error_code text COLLATE "C" NOT NULL,
    evidence_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_profile_backfill_errors_pkey
        PRIMARY KEY (plan_id, page_number, resource_id),
    CONSTRAINT iam_profile_backfill_errors_page_fk FOREIGN KEY (
        plan_id, page_number
    ) REFERENCES iam_profile_backfill_pages (plan_id, page_number)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_errors_code_check CHECK (
        error_code IN (
            'converter_missing', 'source_changed', 'document_invalid',
            'target_schema_invalid', 'checksum_mismatch'
        )
    ),
    CONSTRAINT iam_profile_backfill_errors_digest_check
        CHECK (pg_catalog.octet_length(evidence_sha256) = 32)
);
