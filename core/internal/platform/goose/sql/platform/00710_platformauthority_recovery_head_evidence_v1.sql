-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_recovery_head_evidence_v1(
    in_environment_id text
)
RETURNS TABLE (
    binding_id text,
    environment_id text,
    deployment_mode text,
    registry_kind text,
    license_binding_id text,
    finalization_receipt_sha256 text,
    recovery_trust_bundle_sha256 text,
    signer_generation bigint,
    signer_kid text,
    signer_public_key_sha256 text,
    binding_proof_sha256 text,
    binding_proof_bytes bytea,
    operation_id text,
    purpose text,
    scope_kind text,
    scope_id text,
    generation bigint,
    previous_head_sha256 text,
    state_sha256 text,
    state_bytes bytea,
    proposal_sha256 text,
    proposal_bytes bytea,
    ack_sha256 text,
    ack_bytes bytea,
    registry_sequence bigint,
    completed_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        ''::bytea,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        ''::bytea,
        ''::text,
        ''::bytea,
        ''::text,
        ''::bytea,
        0::bigint,
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_recovery_head_evidence_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_recovery_head_evidence_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_recovery_head_evidence_v1(text);
