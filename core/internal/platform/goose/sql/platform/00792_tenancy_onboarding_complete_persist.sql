-- +goose Up
DROP FUNCTION platform.tenancy_onboarding_complete_v2(uuid, uuid, text, text, bigint);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_complete_v2(
    in_operation_id uuid, in_receipt_id uuid, in_receipt_sha256 text, in_evidence_sha256 text, in_expected_revision bigint
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE; stored_receipt platform.tenancy_onboarding_receipts%ROWTYPE;
BEGIN
    IF in_receipt_sha256 !~ '^[0-9a-f]{64}$' OR in_evidence_sha256 !~ '^[0-9a-f]{64}$' OR in_receipt_id IS NULL THEN
        RAISE EXCEPTION 'activation completion receipt is invalid' USING ERRCODE = '22023';
    END IF;
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'activation operation is missing' USING ERRCODE = '55000'; END IF;
    SELECT * INTO stored_receipt FROM platform.tenancy_onboarding_receipts
    WHERE operation_id = in_operation_id AND receipt_kind = 'activation_completed';
    IF FOUND THEN
        IF stored_receipt.receipt_id = in_receipt_id AND stored_receipt.receipt_sha256 = in_receipt_sha256
           AND stored_receipt.evidence_sha256 = in_evidence_sha256 AND operation.state = 'succeeded' THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION 'activation completion receipt does not match the stored receipt' USING ERRCODE = '23514';
    END IF;
    IF operation.revision <> in_expected_revision OR operation.schema_name IS NULL THEN
        RAISE EXCEPTION 'activation completion revision does not match' USING ERRCODE = '40001';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.tenancy_activation_facts WHERE operation_id = in_operation_id AND tenant_id = operation.tenant_id)
        OR NOT EXISTS (SELECT 1 FROM platform.tenancy_onboarding_receipts WHERE operation_id = in_operation_id AND receipt_kind = 'migration_completed') THEN
        RAISE EXCEPTION 'activation completion is missing its fact or migration receipt' USING ERRCODE = '23514';
    END IF;
    INSERT INTO platform.tenancy_onboarding_receipts (operation_id, receipt_kind, receipt_id, receipt_sha256, evidence_sha256, created_at)
    VALUES (in_operation_id, 'activation_completed', in_receipt_id, in_receipt_sha256, in_evidence_sha256, pg_catalog.clock_timestamp());
    UPDATE platform.tenancy_onboarding_operations SET state = 'succeeded', revision = revision + 1, updated_at = pg_catalog.clock_timestamp()
    WHERE tenancy_onboarding_operations.operation_id = in_operation_id;
    PERFORM platform.tenancy_activation_step_consume_v1(in_operation_id, 'activation.complete', operation.tenant_id, operation.schema_name::name);
    RETURN true;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_complete_v2(uuid, uuid, text, text, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_complete_v2(uuid, uuid, text, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_complete_v2(uuid, uuid, text, text, bigint);
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_complete_v2(in_operation_id uuid, in_receipt_id uuid, in_receipt_sha256 text, in_evidence_sha256 text, in_expected_revision bigint)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_complete_v2(uuid, uuid, text, text, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_complete_v2(uuid, uuid, text, text, bigint) TO harpy_api, harpy_worker;
