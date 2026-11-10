-- +goose Up
DROP FUNCTION platform.platformauthority_development_binding_input_v1(text);

-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_development_binding_input_v1(
    in_environment_id text
)
RETURNS TABLE (
    operation_id text,
    registry_binding_id text,
    finalization_receipt_sha256 text,
    recovery_trust_bundle_sha256 text
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path TO pg_catalog
AS $fn$
DECLARE matched bigint;
BEGIN
    IF in_environment_id IS NULL OR in_environment_id !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN
        RAISE EXCEPTION 'development authority environment identity is invalid' USING ERRCODE = '22023';
    END IF;
    SELECT count(*) INTO matched
    FROM platform.tenancy_onboarding_operations operation
    JOIN platform.tenancy_onboarding_receipts receipt
        ON receipt.operation_id = operation.operation_id
        AND receipt.receipt_kind = 'activation_completed'
    WHERE operation.environment_id = in_environment_id::uuid
        AND operation.state = 'succeeded'
        AND operation.registry_binding_id IS NOT NULL
        AND operation.recovery_trust_bundle_sha256 ~ '^[0-9a-f]{64}$'
        AND receipt.receipt_sha256 ~ '^[0-9a-f]{64}$';
    IF matched IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION 'development authority binding input is missing' USING ERRCODE = '02000';
    END IF;
    RETURN QUERY
    SELECT operation.operation_id::text, operation.registry_binding_id::text,
        receipt.receipt_sha256, operation.recovery_trust_bundle_sha256
    FROM platform.tenancy_onboarding_operations operation
    JOIN platform.tenancy_onboarding_receipts receipt
        ON receipt.operation_id = operation.operation_id
        AND receipt.receipt_kind = 'activation_completed'
    WHERE operation.environment_id = in_environment_id::uuid
        AND operation.state = 'succeeded'
        AND operation.registry_binding_id IS NOT NULL
        AND operation.recovery_trust_bundle_sha256 ~ '^[0-9a-f]{64}$'
        AND receipt.receipt_sha256 ~ '^[0-9a-f]{64}$';
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_development_binding_input_v1(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformauthority_development_binding_input_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_development_binding_input_v1(text);
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_development_binding_input_v1(in_environment_id text)
RETURNS TABLE (operation_id text, registry_binding_id text, finalization_receipt_sha256 text, recovery_trust_bundle_sha256 text)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$ SELECT ''::text, ''::text, ''::text, ''::text; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platformauthority_development_binding_input_v1(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformauthority_development_binding_input_v1(text) TO harpy_api, harpy_worker;
