-- +goose Up
ALTER TABLE iam_password_credentials
    ADD COLUMN legacy_retired_by_job_id uuid,
    DROP CONSTRAINT iam_password_credentials_algorithm_check,
    DROP CONSTRAINT iam_password_credentials_shape_check;

ALTER TABLE iam_password_credentials
    ADD CONSTRAINT iam_password_credentials_algorithm_check CHECK (
        algorithm IN (
            'argon2id_v1', 'bcrypt_v1_adopted', 'legacy_bcrypt_retired_v1'
        )
    ),
    ADD CONSTRAINT iam_password_credentials_shape_check CHECK (
        (algorithm = 'argon2id_v1' AND
            salt IS NOT NULL AND tag IS NOT NULL AND memory_kib IS NOT NULL AND
            iterations IS NOT NULL AND parallelism IS NOT NULL AND
            pg_catalog.octet_length(salt) = 16 AND
            pg_catalog.octet_length(tag) = 32 AND
            memory_kib = 65536 AND iterations = 3 AND parallelism = 4 AND
            legacy_bcrypt_hash IS NULL AND legacy_retired_by_job_id IS NULL AND
            NOT upgrade_required) OR
        (algorithm = 'bcrypt_v1_adopted' AND
            salt IS NULL AND tag IS NULL AND memory_kib IS NULL AND
            iterations IS NULL AND parallelism IS NULL AND
            legacy_bcrypt_hash IS NOT NULL AND legacy_retired_by_job_id IS NULL AND
            pg_catalog.char_length(legacy_bcrypt_hash) = 60 AND
            legacy_bcrypt_hash ~ '^\$2[aby]\$(0[4-9]|[12][0-9]|3[01])\$[./A-Za-z0-9]{53}$' AND
            upgrade_required) OR
        (algorithm = 'legacy_bcrypt_retired_v1' AND
            state = 'reset_required' AND salt IS NULL AND tag IS NULL AND
            memory_kib IS NULL AND iterations IS NULL AND parallelism IS NULL AND
            legacy_bcrypt_hash IS NULL AND legacy_retired_by_job_id IS NOT NULL AND
            upgrade_required)
    );

CREATE INDEX iam_password_credentials_retirement_job_idx
    ON iam_password_credentials (legacy_retired_by_job_id)
    WHERE legacy_retired_by_job_id IS NOT NULL;
