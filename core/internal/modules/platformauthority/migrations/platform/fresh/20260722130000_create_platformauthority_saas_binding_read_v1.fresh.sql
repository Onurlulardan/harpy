-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_saas_binding_read_v1(
    p_environment_id uuid
)
RETURNS TABLE (
    binding_id uuid, environment_id uuid, deployment_mode text,
    registry_kind text, license_binding_id uuid,
    finalization_receipt_sha256 text, recovery_trust_bundle_sha256 text,
    signer_generation bigint, signer_kid text,
    signer_public_key_sha256 text, signer_status text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='SaaS authority binding read is denied';
    END IF;
    RETURN QUERY
    SELECT binding.binding_id,binding.environment_id,
        binding.deployment_mode,binding.registry_kind,
        binding.license_binding_id,binding.finalization_receipt_sha256,
        binding.recovery_trust_bundle_sha256,binding.signer_generation,
        binding.signer_kid,binding.signer_public_key_sha256,
        binding.signer_status
    FROM platform.platformauthority_bindings AS binding
    WHERE binding.environment_id=p_environment_id
      AND binding.deployment_mode='saas'
      AND binding.registry_kind='saas_provider'
      AND binding.signer_status='active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',
            MESSAGE='active SaaS authority binding is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_saas_binding_read_v1(uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformauthority_saas_binding_read_v1(uuid)
TO forgecrud_runtime;
