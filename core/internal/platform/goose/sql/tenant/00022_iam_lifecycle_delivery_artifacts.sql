-- +goose Up
CREATE TABLE iam_lifecycle_delivery_artifacts (
    id uuid NOT NULL PRIMARY KEY,
    subject_user_id uuid NOT NULL,
    resource_id uuid NOT NULL,
    token_id uuid NOT NULL,
    purpose text NOT NULL,
    artifact_ref text NOT NULL,
    kms_key_id uuid NOT NULL,
    nonce text NOT NULL,
    ciphertext text NOT NULL,
    authentication_tag text NOT NULL,
    aad_sha256 text NOT NULL,
    ciphertext_sha256 text NOT NULL,
    expires_at timestamptz,
    recipient_identifier_id uuid NOT NULL
);

-- +goose Down
DROP TABLE iam_lifecycle_delivery_artifacts;
