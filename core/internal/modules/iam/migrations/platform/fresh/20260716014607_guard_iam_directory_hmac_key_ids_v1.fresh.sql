-- +goose Up
ALTER TABLE platform.iam_login_directory
    DROP CONSTRAINT iam_login_directory_lookup_key_check;

ALTER TABLE platform.iam_login_directory
    ADD CONSTRAINT iam_login_directory_lookup_key_check CHECK (
        (pg_catalog.octet_length(lookup_key_id) BETWEEN 1 AND 128 AND
         lookup_key_id ~ '^[a-z0-9][a-z0-9._-]{0,127}$') OR
        platform.iam_hmac_key_id_valid_v1(lookup_key_id)
    );
