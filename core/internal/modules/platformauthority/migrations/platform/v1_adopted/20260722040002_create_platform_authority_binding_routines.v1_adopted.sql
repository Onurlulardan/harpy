-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_bind_v1(
    p_binding_id uuid, p_environment_id uuid, p_deployment_mode text,
    p_registry_kind text, p_license_binding_id uuid,
    p_finalization_receipt_sha256 text, p_recovery_trust_bundle_sha256 text,
    p_signer_generation bigint, p_signer_kid text, p_signer_public_key_sha256 text,
    p_proof_sha256 text, p_proof_bytes bytea, p_created_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_exact boolean;
BEGIN
    INSERT INTO platform.platformauthority_bindings (
        binding_id, environment_id, deployment_mode, registry_kind, license_binding_id,
        finalization_receipt_sha256, recovery_trust_bundle_sha256, signer_generation,
        signer_kid, signer_public_key_sha256, proof_sha256, proof_bytes, created_at
    ) VALUES (
        p_binding_id, p_environment_id, p_deployment_mode, p_registry_kind,
        p_license_binding_id, p_finalization_receipt_sha256,
        p_recovery_trust_bundle_sha256, p_signer_generation, p_signer_kid,
        p_signer_public_key_sha256, p_proof_sha256, p_proof_bytes, p_created_at
    ) ON CONFLICT (binding_id) DO NOTHING;
    IF FOUND THEN
        RETURN;
    END IF;
    SELECT ROW(binding_id, environment_id, deployment_mode, registry_kind,
        license_binding_id, finalization_receipt_sha256, recovery_trust_bundle_sha256,
        signer_generation, signer_kid, signer_public_key_sha256, proof_sha256, proof_bytes)
        IS NOT DISTINCT FROM
        ROW(p_binding_id, p_environment_id, p_deployment_mode, p_registry_kind,
        p_license_binding_id, p_finalization_receipt_sha256, p_recovery_trust_bundle_sha256,
        p_signer_generation, p_signer_kid, p_signer_public_key_sha256,
        p_proof_sha256, p_proof_bytes)
    INTO v_exact
    FROM platform.platformauthority_bindings
    WHERE binding_id = p_binding_id;
    IF NOT COALESCE(v_exact, false) THEN
        RAISE EXCEPTION USING ERRCODE = '23505', MESSAGE = 'protected authority binding conflicts';
    END IF;
EXCEPTION WHEN unique_violation THEN
    RAISE EXCEPTION USING ERRCODE = '23505', MESSAGE = 'protected authority environment cannot be rebound';
END;
$function$;

CREATE FUNCTION platform.platformauthority_get_binding_v1(p_binding_id uuid)
RETURNS TABLE (
    binding_id uuid, environment_id uuid, deployment_mode text, registry_kind text,
    signer_generation bigint, signer_kid text, signer_public_key_sha256 text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY SELECT binding.binding_id, binding.environment_id,
        binding.deployment_mode, binding.registry_kind, binding.signer_generation,
        binding.signer_kid, binding.signer_public_key_sha256
    FROM platform.platformauthority_bindings AS binding
    WHERE binding.binding_id = p_binding_id;
END;
$function$;
-- +goose StatementEnd
