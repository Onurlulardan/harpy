-- +goose Up
ALTER TABLE platform.iam_auth_throttle_buckets
    DROP CONSTRAINT iam_auth_throttle_buckets_key_check,
    ADD CONSTRAINT iam_auth_throttle_buckets_key_check CHECK (
        platform.iam_hmac_key_id_valid_v1(key_id) AND
        pg_catalog.octet_length(bucket_hmac) = 32
    );
