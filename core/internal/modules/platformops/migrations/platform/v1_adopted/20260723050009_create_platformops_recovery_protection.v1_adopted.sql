-- +goose Up
CREATE TABLE platform.platformops_recovery_protections (
    environment_id uuid PRIMARY KEY,
    state text NOT NULL CHECK (state IN (
        'unprotected','rotation_pending','disable_pending','protected'
    )),
    generation bigint NOT NULL CHECK (generation >= 0),
    current_operation_id uuid,
    active_kit_sha256 bytea CHECK (
        active_kit_sha256 IS NULL OR pg_catalog.octet_length(active_kit_sha256) = 32
    ),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    CONSTRAINT platformops_recovery_protection_identity_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.platformops_installation_recovery_identities(environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT platformops_recovery_protection_state_ck CHECK (
        (state = 'unprotected' AND active_kit_sha256 IS NULL) OR
        (state = 'protected' AND generation > 0 AND active_kit_sha256 IS NOT NULL) OR
        state IN ('rotation_pending','disable_pending')
    )
);

CREATE TABLE platform.platformops_recovery_control_operations (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    operation_kind text NOT NULL CHECK (operation_kind IN ('enable','rotate','disable')),
    actor_user_id uuid NOT NULL,
    idempotency_key text COLLATE "C" NOT NULL CHECK (
        idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$'
    ),
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    claim_algorithm text,
    claim_public_key bytea,
    state text NOT NULL CHECK (state IN (
        'accepted','handoff_ready','kit_ready','completed','failed','compensated'
    )),
    kit_claim_state text NOT NULL CHECK (kit_claim_state IN (
        'not_applicable','pending','ready','downloaded','acknowledged','expired'
    )),
    artifact_sha256 bytea CHECK (
        artifact_sha256 IS NULL OR pg_catalog.octet_length(artifact_sha256) = 32
    ),
    artifact_bytes bytea,
    ack_verifier_sha256 bytea CHECK (
        ack_verifier_sha256 IS NULL OR pg_catalog.octet_length(ack_verifier_sha256) = 32
    ),
    claimed_session_id uuid,
    expires_at timestamptz,
    failure_code text,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    completed_at timestamptz,
    CONSTRAINT platformops_recovery_control_environment_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.platformops_recovery_protections(environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT platformops_recovery_control_idempotency_uq
        UNIQUE (environment_id,operation_kind,idempotency_key),
    CONSTRAINT platformops_recovery_control_environment_operation_uq
        UNIQUE (environment_id,operation_id),
    CONSTRAINT platformops_recovery_control_claim_ck CHECK (
        (operation_kind = 'disable' AND claim_algorithm IS NULL AND
         claim_public_key IS NULL AND kit_claim_state = 'not_applicable') OR
        (operation_kind IN ('enable','rotate') AND
         claim_algorithm = 'HPKE-v1-X25519-HKDF-SHA256-AES-256-GCM' AND
         pg_catalog.octet_length(claim_public_key) = 32 AND
         kit_claim_state <> 'not_applicable')
    ),
    CONSTRAINT platformops_recovery_control_artifact_ck CHECK (
        (artifact_bytes IS NULL AND artifact_sha256 IS NULL AND
         ack_verifier_sha256 IS NULL) OR
        (artifact_bytes IS NOT NULL AND artifact_sha256 IS NOT NULL AND
         ack_verifier_sha256 IS NOT NULL) OR
        (kit_claim_state = 'acknowledged' AND artifact_bytes IS NULL AND
         artifact_sha256 IS NOT NULL AND ack_verifier_sha256 IS NULL)
    )
);

ALTER TABLE platform.platformops_recovery_protections
ADD CONSTRAINT platformops_recovery_protection_operation_fk
FOREIGN KEY (environment_id,current_operation_id)
REFERENCES platform.platformops_recovery_control_operations(environment_id,operation_id)
ON UPDATE NO ACTION ON DELETE RESTRICT;

INSERT INTO platform.platformops_recovery_protections(
    environment_id,state,generation,created_at,updated_at
)
SELECT identity.environment_id,'unprotected',0,identity.created_at,identity.created_at
FROM platform.platformops_installation_recovery_identities AS identity
WHERE identity.deployment_mode = 'onpremise'
ON CONFLICT (environment_id) DO NOTHING;

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_initialize_recovery_protection_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NEW.deployment_mode = 'onpremise' THEN
        INSERT INTO platform.platformops_recovery_protections(
            environment_id,state,generation,created_at,updated_at
        ) VALUES (NEW.environment_id,'unprotected',0,NEW.created_at,NEW.created_at)
        ON CONFLICT (environment_id) DO NOTHING;
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER platformops_installation_recovery_protection_initialize
AFTER INSERT ON platform.platformops_installation_recovery_identities
FOR EACH ROW EXECUTE FUNCTION platform.platformops_initialize_recovery_protection_v1();
