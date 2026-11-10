-- +goose Up
DROP FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_activate_v3(in_operation_id uuid, in_lease_token_sha256 text, in_expected_revision bigint)
RETURNS TABLE (operation_id uuid, tenant_id uuid, operation_kind text, step_plan_sha256 text, migration_receipt_id uuid, migration_receipt_sha256 text, authority_ack_sha256 text, revision bigint, database_now timestamptz, lease_until timestamptz)
LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE
    operation platform.tenancy_onboarding_operations%ROWTYPE;
    lease platform.tenancy_onboarding_leases%ROWTYPE;
    payload text;
    receipt uuid;
    receipt_sha text;
    observed timestamptz;
    command_name text;
    step_ordinal integer := 0;
BEGIN
    IF session_user NOT IN ('harpy_api', 'harpy_worker', 'harpy_owner', 'harpy_v2_runtime') THEN
        RAISE EXCEPTION 'activation role is not authorized' USING ERRCODE = '42501';
    END IF;
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations AS stored WHERE stored.operation_id = in_operation_id FOR UPDATE;
    IF NOT FOUND OR operation.revision <> in_expected_revision OR operation.schema_name IS NULL OR operation.manifest_sha256 IS NULL THEN
        RAISE EXCEPTION 'activation operation is not ready' USING ERRCODE = '55000';
    END IF;
    SELECT * INTO lease FROM platform.tenancy_onboarding_leases AS stored
    WHERE stored.operation_id = in_operation_id AND stored.tenant_id = operation.tenant_id AND stored.lease_token_sha256 = in_lease_token_sha256;
    IF NOT FOUND OR lease.lease_until <= pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION 'activation lease is not valid' USING ERRCODE = '55P03';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_namespace WHERE nspname = operation.schema_name) THEN
        RAISE EXCEPTION 'tenant schema is not provisioned' USING ERRCODE = '55000';
    END IF;
    payload := 'tenancy.migration_completed|' || operation.operation_id::text || '|' || operation.tenant_id::text || '|'
        || operation.schema_name || '|' || operation.step_plan_sha256 || '|' || operation.manifest_sha256;
    receipt := platform.tenancy_activation_digest_uuid_v1(payload);
    receipt_sha := platform.tenancy_activation_digest_hex_v1(payload);
    INSERT INTO platform.tenancy_onboarding_receipts (operation_id, receipt_kind, receipt_id, receipt_sha256, evidence_sha256, created_at)
    VALUES (operation.operation_id, 'migration_completed', receipt, receipt_sha, receipt_sha, pg_catalog.clock_timestamp())
    ON CONFLICT ON CONSTRAINT tenancy_onboarding_receipts_pkey DO NOTHING;
    SELECT stored.receipt_id, stored.receipt_sha256 INTO receipt, receipt_sha
    FROM platform.tenancy_onboarding_receipts AS stored
    WHERE stored.operation_id = operation.operation_id AND stored.receipt_kind = 'migration_completed';
    IF receipt IS NULL OR receipt_sha <> platform.tenancy_activation_digest_hex_v1(payload) THEN
        RAISE EXCEPTION 'migration receipt does not match the provisioned schema' USING ERRCODE = '23514';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.tenancy_onboarding_receipts AS stored
        WHERE stored.operation_id = operation.operation_id AND stored.receipt_kind = 'activation_completed') THEN
    FOREACH command_name IN ARRAY ARRAY[
        'tenancy.activate_tenant', 'iam.initialize_tenant_identity', 'billing.assign_tenant_base_offering',
        'billing.open_tenant_usage',
        'platformops.initialize_recovery_signing_identity', 'platformauthority.bind_owner',
        'platformauthority.pin_protected_authority_head', 'activation.complete']
    LOOP
        IF operation.kind = 'saas_customer_tenant_provision' AND command_name IN ('platformops.initialize_recovery_signing_identity', 'platformauthority.bind_owner') THEN CONTINUE; END IF;
        IF operation.deployment_mode = 'onpremise' AND command_name = 'platformauthority.pin_protected_authority_head' THEN CONTINUE; END IF;
        step_ordinal := step_ordinal + 1;
        INSERT INTO platform.tenancy_activation_step_capabilities (
            operation_id, ordinal, command_name, tenant_id, environment_id, profile_kind, manifest_sha256,
            schema_name, session_user_name, backend_pid, transaction_id, created_at
        ) VALUES (
            operation.operation_id, step_ordinal, command_name, operation.tenant_id, operation.environment_id,
            operation.kind, operation.manifest_sha256, operation.schema_name, session_user,
            pg_catalog.pg_backend_pid(), pg_catalog.pg_current_xact_id(), pg_catalog.clock_timestamp()
        );
    END LOOP;
    END IF;
    observed := pg_catalog.clock_timestamp();
    RETURN QUERY SELECT operation.operation_id, operation.tenant_id, operation.kind, operation.step_plan_sha256,
        receipt, receipt_sha,
        (SELECT stored.ack_sha256 FROM platform.tenancy_onboarding_authority_acks AS stored WHERE stored.operation_id = operation.operation_id),
        operation.revision, observed, lease.lease_until;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint);
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_activate_v3(in_operation_id uuid, in_lease_token_sha256 text, in_expected_revision bigint)
RETURNS TABLE (operation_id uuid, tenant_id uuid, operation_kind text, step_plan_sha256 text, migration_receipt_id uuid, migration_receipt_sha256 text, authority_ack_sha256 text, revision bigint, database_now timestamptz, lease_until timestamptz)
LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE
    operation platform.tenancy_onboarding_operations%ROWTYPE;
    lease platform.tenancy_onboarding_leases%ROWTYPE;
    payload text;
    receipt uuid;
    receipt_sha text;
    observed timestamptz;
    command_name text;
    step_ordinal integer := 0;
BEGIN
    IF session_user NOT IN ('harpy_api', 'harpy_worker', 'harpy_owner', 'harpy_v2_runtime') THEN
        RAISE EXCEPTION 'activation role is not authorized' USING ERRCODE = '42501';
    END IF;
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations AS stored WHERE stored.operation_id = in_operation_id FOR UPDATE;
    IF NOT FOUND OR operation.revision <> in_expected_revision OR operation.schema_name IS NULL OR operation.manifest_sha256 IS NULL THEN
        RAISE EXCEPTION 'activation operation is not ready' USING ERRCODE = '55000';
    END IF;
    SELECT * INTO lease FROM platform.tenancy_onboarding_leases AS stored
    WHERE stored.operation_id = in_operation_id AND stored.tenant_id = operation.tenant_id AND stored.lease_token_sha256 = in_lease_token_sha256;
    IF NOT FOUND OR lease.lease_until <= pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION 'activation lease is not valid' USING ERRCODE = '55P03';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_namespace WHERE nspname = operation.schema_name) THEN
        RAISE EXCEPTION 'tenant schema is not provisioned' USING ERRCODE = '55000';
    END IF;
    payload := 'tenancy.migration_completed|' || operation.operation_id::text || '|' || operation.tenant_id::text || '|'
        || operation.schema_name || '|' || operation.step_plan_sha256 || '|' || operation.manifest_sha256;
    receipt := platform.tenancy_activation_digest_uuid_v1(payload);
    receipt_sha := platform.tenancy_activation_digest_hex_v1(payload);
    INSERT INTO platform.tenancy_onboarding_receipts (operation_id, receipt_kind, receipt_id, receipt_sha256, evidence_sha256, created_at)
    VALUES (operation.operation_id, 'migration_completed', receipt, receipt_sha, receipt_sha, pg_catalog.clock_timestamp())
    ON CONFLICT ON CONSTRAINT tenancy_onboarding_receipts_pkey DO NOTHING;
    SELECT stored.receipt_id, stored.receipt_sha256 INTO receipt, receipt_sha
    FROM platform.tenancy_onboarding_receipts AS stored
    WHERE stored.operation_id = operation.operation_id AND stored.receipt_kind = 'migration_completed';
    IF receipt IS NULL OR receipt_sha <> platform.tenancy_activation_digest_hex_v1(payload) THEN
        RAISE EXCEPTION 'migration receipt does not match the provisioned schema' USING ERRCODE = '23514';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.tenancy_onboarding_receipts AS stored
        WHERE stored.operation_id = operation.operation_id AND stored.receipt_kind = 'activation_completed') THEN
    FOREACH command_name IN ARRAY ARRAY[
        'tenancy.activate_tenant', 'iam.initialize_tenant_identity', 'billing.assign_tenant_base_offering',
        'billing.open_tenant_usage', 'systemresources.initialize_tenant_baseline',
        'platformops.initialize_recovery_signing_identity', 'platformauthority.bind_owner',
        'platformauthority.pin_protected_authority_head', 'activation.complete']
    LOOP
        IF operation.kind = 'saas_customer_tenant_provision' AND command_name IN ('platformops.initialize_recovery_signing_identity', 'platformauthority.bind_owner') THEN CONTINUE; END IF;
        IF operation.deployment_mode = 'onpremise' AND command_name = 'platformauthority.pin_protected_authority_head' THEN CONTINUE; END IF;
        step_ordinal := step_ordinal + 1;
        INSERT INTO platform.tenancy_activation_step_capabilities (
            operation_id, ordinal, command_name, tenant_id, environment_id, profile_kind, manifest_sha256,
            schema_name, session_user_name, backend_pid, transaction_id, created_at
        ) VALUES (
            operation.operation_id, step_ordinal, command_name, operation.tenant_id, operation.environment_id,
            operation.kind, operation.manifest_sha256, operation.schema_name, session_user,
            pg_catalog.pg_backend_pid(), pg_catalog.pg_current_xact_id(), pg_catalog.clock_timestamp()
        );
    END LOOP;
    END IF;
    observed := pg_catalog.clock_timestamp();
    RETURN QUERY SELECT operation.operation_id, operation.tenant_id, operation.kind, operation.step_plan_sha256,
        receipt, receipt_sha,
        (SELECT stored.ack_sha256 FROM platform.tenancy_onboarding_authority_acks AS stored WHERE stored.operation_id = operation.operation_id),
        operation.revision, observed, lease.lease_until;
END;
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint) TO harpy_api, harpy_worker;
