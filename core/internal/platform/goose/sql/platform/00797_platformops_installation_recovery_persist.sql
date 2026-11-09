-- +goose Up
CREATE TABLE IF NOT EXISTS platform.platformops_installation_recovery_identities (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    schema_name name NOT NULL,
    deployment_mode text NOT NULL,
    signing_generation bigint,
    trust_bundle_sha256 text,
    manifest_sha256 text NOT NULL,
    created_at timestamptz NOT NULL
);

REVOKE ALL ON TABLE platform.platformops_installation_recovery_identities FROM PUBLIC, harpy_api, harpy_worker;

DROP FUNCTION platform.platformops_installation_recovery_initialize_v2(uuid, uuid, uuid, name, text, bigint, text, text);

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_installation_recovery_initialize_v2(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name, in_deployment_mode text,
    in_signing_generation bigint, in_trust_bundle_sha256 text, in_manifest_sha256 text
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE;
BEGIN
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id;
    IF NOT FOUND OR operation.tenant_id <> in_tenant_id OR operation.environment_id IS DISTINCT FROM in_environment_id
        OR operation.schema_name IS DISTINCT FROM in_schema_name::text OR operation.deployment_mode IS DISTINCT FROM in_deployment_mode
        OR operation.manifest_sha256 IS DISTINCT FROM in_manifest_sha256
        OR operation.recovery_generation IS DISTINCT FROM in_signing_generation
        OR operation.recovery_trust_bundle_sha256 IS DISTINCT FROM in_trust_bundle_sha256
        OR NOT platform.tenancy_activation_assert_v2(in_operation_id, in_tenant_id, in_schema_name) THEN
        RAISE EXCEPTION 'recovery signing identity does not match the activation' USING ERRCODE = '42501';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.platformops_installation_recovery_identities WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id
        AND signing_generation IS NOT DISTINCT FROM in_signing_generation AND trust_bundle_sha256 IS NOT DISTINCT FROM in_trust_bundle_sha256) THEN
        RETURN true;
    END IF;
    INSERT INTO platform.platformops_installation_recovery_identities (
        operation_id, environment_id, tenant_id, schema_name, deployment_mode, signing_generation, trust_bundle_sha256, manifest_sha256, created_at
    ) VALUES (
        in_operation_id, in_environment_id, in_tenant_id, in_schema_name, in_deployment_mode, in_signing_generation,
        in_trust_bundle_sha256, in_manifest_sha256, pg_catalog.clock_timestamp()
    );
    PERFORM platform.tenancy_activation_step_consume_v1(in_operation_id, 'platformops.initialize_recovery_signing_identity', in_tenant_id, in_schema_name);
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_installation_recovery_initialize_v2(uuid, uuid, uuid, name, text, bigint, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_installation_recovery_initialize_v2(uuid, uuid, uuid, name, text, bigint, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_installation_recovery_initialize_v2(uuid, uuid, uuid, name, text, bigint, text, text);
DROP TABLE IF EXISTS platform.platformops_installation_recovery_identities;
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_installation_recovery_initialize_v2(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name, in_deployment_mode text, in_signing_generation bigint, in_trust_bundle_sha256 text, in_manifest_sha256 text)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platformops_installation_recovery_initialize_v2(uuid, uuid, uuid, name, text, bigint, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_installation_recovery_initialize_v2(uuid, uuid, uuid, name, text, bigint, text, text) TO harpy_api, harpy_worker;
