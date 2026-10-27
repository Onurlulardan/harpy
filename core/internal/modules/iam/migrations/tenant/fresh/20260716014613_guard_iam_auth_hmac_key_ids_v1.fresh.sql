-- +goose Up
ALTER TABLE iam_sessions DROP CONSTRAINT iam_sessions_ip_hmac_check,
    ADD CONSTRAINT iam_sessions_ip_hmac_check CHECK (
        iam_hmac_key_id_valid_v1(ip_hmac_key_id) AND
        pg_catalog.octet_length(ip_hmac) = 32
    );
ALTER TABLE iam_recovery_codes DROP CONSTRAINT iam_recovery_codes_hmac_key_id_check,
    ADD CONSTRAINT iam_recovery_codes_hmac_key_id_check
        CHECK (iam_hmac_key_id_valid_v1(hmac_key_id));
ALTER TABLE iam_auth_challenges DROP CONSTRAINT iam_auth_challenges_secret_check,
    ADD CONSTRAINT iam_auth_challenges_secret_check CHECK (
        iam_hmac_key_id_valid_v1(hmac_key_id) AND
        pg_catalog.octet_length(challenge_hmac) = 32 AND
        pg_catalog.octet_length(context_sha256) = 32
    );
ALTER TABLE iam_one_time_tokens DROP CONSTRAINT iam_one_time_tokens_secret_check,
    ADD CONSTRAINT iam_one_time_tokens_secret_check CHECK (
        pg_catalog.octet_length(selector) = 16 AND
        iam_hmac_key_id_valid_v1(hash_key_id) AND
        pg_catalog.octet_length(secret_hmac) = 32 AND
        pg_catalog.octet_length(binding_sha256) = 32
    );
ALTER TABLE iam_refresh_tokens DROP CONSTRAINT iam_refresh_tokens_secret_check,
    ADD CONSTRAINT iam_refresh_tokens_secret_check CHECK (
        pg_catalog.octet_length(selector) = 16 AND
        iam_hmac_key_id_valid_v1(hash_key_id) AND
        pg_catalog.octet_length(secret_hmac) = 32
    );
ALTER TABLE iam_totp_authenticators
    DROP CONSTRAINT iam_totp_authenticators_kms_key_id_check,
    ADD CONSTRAINT iam_totp_authenticators_kms_key_id_check CHECK (
        pg_catalog.octet_length(kms_key_id) BETWEEN 1 AND 512 AND
        kms_key_id = pg_catalog.btrim(kms_key_id) AND
        kms_key_id !~ '[[:cntrl:]]'
    );
