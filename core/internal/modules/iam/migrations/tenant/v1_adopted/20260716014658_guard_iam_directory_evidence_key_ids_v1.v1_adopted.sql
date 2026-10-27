-- +goose Up
ALTER TABLE iam_directory_key_evidence
    DROP CONSTRAINT iam_directory_key_evidence_key_check;

ALTER TABLE iam_directory_key_evidence
    ADD CONSTRAINT iam_directory_key_evidence_key_check CHECK (
        (pg_catalog.octet_length(lookup_key_id) BETWEEN 1 AND 128 AND
         lookup_key_id ~ '^[a-z0-9][a-z0-9._-]{0,127}$') OR
        iam_hmac_key_id_valid_v1(lookup_key_id)
    );
