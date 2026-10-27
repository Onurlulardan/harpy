-- +goose Up
ALTER TABLE iam_totp_authenticators
    DROP CONSTRAINT iam_totp_authenticators_kms_key_id_check;

ALTER TABLE iam_totp_authenticators
    ADD CONSTRAINT iam_totp_authenticators_kms_key_id_check CHECK (
        pg_catalog.octet_length(kms_key_id) BETWEEN 1 AND 512 AND
        kms_key_id ~ '^gcp-kms://projects/([a-z][a-z0-9-]{4,28}[a-z0-9]|[0-9]{6,30})/locations/[a-z0-9][a-z0-9-]{0,62}/keyRings/[A-Za-z0-9_-]{1,63}/cryptoKeys/[A-Za-z0-9_-]{1,63}/cryptoKeyVersions/[1-9][0-9]{0,18}$'
    );
