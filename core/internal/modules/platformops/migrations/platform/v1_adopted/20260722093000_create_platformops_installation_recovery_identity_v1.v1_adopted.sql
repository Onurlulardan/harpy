-- +goose Up
CREATE TABLE platform.platformops_installation_recovery_identities (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL UNIQUE,
    tenant_id uuid NOT NULL UNIQUE,
    deployment_mode text NOT NULL
        CHECK (deployment_mode IN ('onpremise', 'saas')),
    identity_state text NOT NULL
        CHECK (identity_state IN ('active', 'not_applicable')),
    signing_generation bigint,
    trust_bundle_sha256 text,
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL,
    CONSTRAINT platformops_installation_recovery_identity_state_check CHECK (
        (deployment_mode = 'onpremise' AND identity_state = 'active'
            AND signing_generation > 0
            AND trust_bundle_sha256 ~ '^[0-9a-f]{64}$')
        OR
        (deployment_mode = 'saas' AND identity_state = 'not_applicable'
            AND signing_generation IS NULL
            AND trust_bundle_sha256 IS NULL)
    )
);

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_reject_installation_identity_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'installation recovery identity is immutable';
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER platformops_installation_recovery_identity_immutable
BEFORE UPDATE OR DELETE
ON platform.platformops_installation_recovery_identities
FOR EACH ROW
EXECUTE FUNCTION platform.platformops_reject_installation_identity_mutation_v1();
