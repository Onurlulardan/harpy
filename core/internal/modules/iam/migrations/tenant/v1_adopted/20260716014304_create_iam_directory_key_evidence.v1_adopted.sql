-- +goose Up
CREATE TABLE iam_directory_key_evidence (
    lookup_key_id text COLLATE "C" NOT NULL,
    last_job_id uuid NOT NULL,
    projected_row_count bigint NOT NULL,
    zero_rows_since timestamptz,
    last_result_sha256 bytea NOT NULL,
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_directory_key_evidence_pkey PRIMARY KEY (lookup_key_id),
    CONSTRAINT iam_directory_key_evidence_key_check CHECK (
        pg_catalog.octet_length(lookup_key_id) BETWEEN 1 AND 128 AND
        lookup_key_id ~ '^[a-z0-9][a-z0-9._-]{0,127}$'
    ),
    CONSTRAINT iam_directory_key_evidence_count_check CHECK (
        projected_row_count >= 0 AND revision > 0 AND
        ((projected_row_count = 0 AND zero_rows_since IS NOT NULL) OR
         (projected_row_count > 0 AND zero_rows_since IS NULL))
    ),
    CONSTRAINT iam_directory_key_evidence_digest_check
        CHECK (pg_catalog.octet_length(last_result_sha256) = 32)
);
