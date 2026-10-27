-- +goose Up
ALTER TABLE platform.iam_auth_throttle_buckets
    ADD COLUMN authority_epoch bigint NOT NULL;

ALTER TABLE platform.iam_auth_throttle_buckets
    DROP CONSTRAINT iam_auth_throttle_buckets_pkey,
    ADD CONSTRAINT iam_auth_throttle_buckets_pkey PRIMARY KEY (
        bucket_scope_kind, scope_id, purpose, dimension,
        key_id, bucket_hmac, authority_epoch
    ),
    ADD CONSTRAINT iam_auth_throttle_buckets_authority_epoch_check CHECK (
        (
            bucket_scope_kind = 'tenant' AND
            authority_epoch = authority_group_generation AND
            authority_epoch = cutover_generation
        ) OR (
            bucket_scope_kind = 'environment' AND authority_epoch = 0
        )
    );
