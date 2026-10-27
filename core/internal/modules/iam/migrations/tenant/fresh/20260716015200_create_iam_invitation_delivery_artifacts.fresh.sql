-- +goose Up
CREATE TABLE iam_invitation_delivery_artifacts (
    id uuid NOT NULL,
    invitation_id uuid NOT NULL,
    token_id uuid NOT NULL,
    artifact_ref text COLLATE "C" NOT NULL,
    kms_key_id text COLLATE "C" NOT NULL,
    nonce bytea NOT NULL,
    ciphertext bytea NOT NULL,
    authentication_tag bytea NOT NULL,
    aad_sha256 bytea NOT NULL,
    ciphertext_sha256 bytea NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    CONSTRAINT iam_invitation_delivery_artifacts_pkey PRIMARY KEY (id),
    CONSTRAINT iam_invitation_delivery_artifacts_invitation_key
        UNIQUE (invitation_id),
    CONSTRAINT iam_invitation_delivery_artifacts_token_key UNIQUE (token_id),
    CONSTRAINT iam_invitation_delivery_artifacts_ref_key UNIQUE (artifact_ref),
    CONSTRAINT iam_invitation_delivery_artifacts_invitation_fk
        FOREIGN KEY (invitation_id) REFERENCES iam_invitations (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitation_delivery_artifacts_token_fk
        FOREIGN KEY (token_id) REFERENCES iam_one_time_tokens (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitation_delivery_artifacts_ref_check CHECK (
        artifact_ref ~ '^iam-delivery://tenant/[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}/artifact/[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
    ),
    CONSTRAINT iam_invitation_delivery_artifacts_kms_check CHECK (
        pg_catalog.char_length(kms_key_id) <= 512 AND kms_key_id ~
        '^gcp-kms://projects/([a-z][a-z0-9-]{4,28}[a-z0-9]|[0-9]{6,30})/locations/[a-z0-9][a-z0-9-]{0,62}/keyRings/[A-Za-z0-9_-]{1,63}/cryptoKeys/[A-Za-z0-9_-]{1,63}/cryptoKeyVersions/[1-9][0-9]{0,18}$'
    ),
    CONSTRAINT iam_invitation_delivery_artifacts_envelope_check CHECK (
        pg_catalog.octet_length(nonce) = 12 AND
        pg_catalog.octet_length(ciphertext) = 66 AND
        pg_catalog.octet_length(authentication_tag) = 16 AND
        pg_catalog.octet_length(aad_sha256) = 32 AND
        pg_catalog.octet_length(ciphertext_sha256) = 32
    ),
    CONSTRAINT iam_invitation_delivery_artifacts_lifecycle_check CHECK (
        revision = 1 AND expires_at > created_at
    )
);

CREATE TRIGGER iam_invitation_delivery_artifacts_immutable_guard
BEFORE UPDATE OR TRUNCATE ON iam_invitation_delivery_artifacts
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_invitation_delivery_artifacts ENABLE ALWAYS TRIGGER
    iam_invitation_delivery_artifacts_immutable_guard;
