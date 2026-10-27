-- +goose Up
CREATE TABLE iam_invitation_delivery_worker_claims (
    artifact_id uuid NOT NULL,
    artifact_ref text COLLATE "C" NOT NULL,
    effect_id uuid NOT NULL,
    effect_idempotency_key text COLLATE "C" NOT NULL,
    message_id text COLLATE "C" NOT NULL,
    state text COLLATE "C" NOT NULL,
    claimed_at timestamptz NOT NULL,
    outcome_unknown_at timestamptz,
    CONSTRAINT iam_invitation_delivery_worker_claims_pkey
        PRIMARY KEY (artifact_id),
    CONSTRAINT iam_invitation_delivery_worker_claims_ref_key
        UNIQUE (artifact_ref),
    CONSTRAINT iam_invitation_delivery_worker_claims_effect_key
        UNIQUE (effect_id),
    CONSTRAINT iam_invitation_delivery_worker_claims_message_key
        UNIQUE (message_id),
    CONSTRAINT iam_invitation_delivery_worker_claims_artifact_fk
        FOREIGN KEY (artifact_id)
        REFERENCES iam_invitation_delivery_artifacts (id)
        ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT iam_invitation_delivery_worker_claims_identity_check CHECK (
        pg_catalog.char_length(effect_idempotency_key) BETWEEN 1 AND 512 AND
        pg_catalog.char_length(message_id) BETWEEN 3 AND 512
    ),
    CONSTRAINT iam_invitation_delivery_worker_claims_state_check CHECK (
        (state = 'claimed' AND outcome_unknown_at IS NULL) OR
        (state = 'outcome_unknown' AND outcome_unknown_at IS NOT NULL AND
         outcome_unknown_at >= claimed_at)
    )
);

CREATE INDEX iam_invitation_delivery_worker_claims_state_idx
ON iam_invitation_delivery_worker_claims (state, claimed_at);

GRANT SELECT ON TABLE iam_invitation_delivery_worker_claims
TO harpy_runtime;
