-- +goose Up
CREATE TABLE platform.platformauthority_signer_events (
    event_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    binding_id uuid NOT NULL REFERENCES platform.platformauthority_bindings (binding_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    signer_generation bigint NOT NULL CHECK (signer_generation > 0),
    signer_kid text NOT NULL CHECK (octet_length(signer_kid) BETWEEN 1 AND 512),
    signer_public_key_sha256 text NOT NULL CHECK (signer_public_key_sha256 ~ '^[0-9a-f]{64}$'),
    event_kind text NOT NULL CHECK (event_kind IN ('activated', 'retired', 'revoked')),
    evidence_sha256 text CHECK (evidence_sha256 ~ '^[0-9a-f]{64}$'),
    occurred_at timestamptz NOT NULL,
    UNIQUE (binding_id, signer_generation, event_kind)
);

INSERT INTO platform.platformauthority_signer_events (
    binding_id, signer_generation, signer_kid, signer_public_key_sha256,
    event_kind, occurred_at
)
SELECT binding_id, signer_generation, signer_kid, signer_public_key_sha256,
    'activated', created_at
FROM platform.platformauthority_bindings;

CREATE TRIGGER platformauthority_signer_events_guard
BEFORE UPDATE OR DELETE ON platform.platformauthority_signer_events
FOR EACH ROW EXECUTE FUNCTION platform.platformauthority_reject_mutation_v1();
