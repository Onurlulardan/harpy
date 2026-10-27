-- +goose Up
CREATE TABLE iam_invitation_delivery_worker_receipts (
    receipt_id uuid NOT NULL,
    artifact_id uuid NOT NULL,
    invitation_id uuid NOT NULL,
    token_id uuid NOT NULL,
    artifact_ref text COLLATE "C" NOT NULL,
    effect_id uuid NOT NULL,
    effect_idempotency_key text COLLATE "C" NOT NULL,
    message_id text COLLATE "C" NOT NULL,
    outcome text COLLATE "C" NOT NULL,
    code text COLLATE "C" NOT NULL,
    provider_reference text COLLATE "C" NOT NULL,
    result_sha256 bytea NOT NULL,
    completed_at timestamptz NOT NULL,
    CONSTRAINT iam_invitation_delivery_worker_receipts_pkey
        PRIMARY KEY (receipt_id),
    CONSTRAINT iam_invitation_delivery_worker_receipts_artifact_key
        UNIQUE (artifact_id),
    CONSTRAINT iam_invitation_delivery_worker_receipts_ref_key
        UNIQUE (artifact_ref),
    CONSTRAINT iam_invitation_delivery_worker_receipts_effect_key
        UNIQUE (effect_id),
    CONSTRAINT iam_invitation_delivery_worker_receipts_idempotency_key
        UNIQUE (effect_idempotency_key),
    CONSTRAINT iam_invitation_delivery_worker_receipts_message_key
        UNIQUE (message_id),
    CONSTRAINT iam_invitation_delivery_worker_receipts_identity_check CHECK (
        pg_catalog.char_length(effect_idempotency_key) BETWEEN 1 AND 512 AND
        pg_catalog.char_length(message_id) BETWEEN 3 AND 512 AND
        pg_catalog.char_length(provider_reference) BETWEEN 1 AND 512 AND
        code ~ '^[A-Z][A-Z0-9_]{1,63}$' AND
        pg_catalog.octet_length(result_sha256) = 32
    ),
    CONSTRAINT iam_invitation_delivery_worker_receipts_outcome_check CHECK (
        outcome IN ('delivered', 'permanent_failure')
    )
);

CREATE TRIGGER iam_invitation_delivery_worker_receipts_immutable_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_invitation_delivery_worker_receipts
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_invitation_delivery_worker_receipts ENABLE ALWAYS TRIGGER
    iam_invitation_delivery_worker_receipts_immutable_guard;

GRANT SELECT
ON TABLE iam_invitation_delivery_worker_receipts TO forgecrud_runtime;
