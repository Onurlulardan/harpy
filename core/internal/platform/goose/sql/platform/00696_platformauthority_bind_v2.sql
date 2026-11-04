-- +goose Up
CREATE FUNCTION platform.platformauthority_bind_v2(
    in_binding_id text,
    in_environment_id text,
    in_deployment_mode text,
    in_registry_kind text,
    in_license_binding_id text,
    in_finalization_receipt_sha256 text,
    in_recovery_trust_bundle_sha256 text,
    in_signer_generation bigint,
    in_signer_kid text,
    in_signer_public_key_sha256 text,
    in_proof_sha text,
    in_proof_bytes bytea,
    in_now timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_bind_v2(text, text, text, text, text, text, text, bigint, text, text, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_bind_v2(text, text, text, text, text, text, text, bigint, text, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_bind_v2(text, text, text, text, text, text, text, bigint, text, text, text, bytea, timestamptz);
