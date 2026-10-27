-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_stage_v1(
    p_operation_id uuid,
    p_attempt_id uuid,
    p_credential_id uuid,
    p_credential_version bigint,
    p_installation_manifest_sha256 text,
    p_delivery_manifest_sha256 text
)
RETURNS TABLE (
    operation_id uuid, tenant_id uuid, first_owner_user_id uuid,
    current_attempt_id uuid, current_credential_id uuid,
    current_credential_version bigint, delivery_state text,
    delivery_revision bigint, database_now timestamptz,
    claim_adapter_id text, claim_lease_until timestamptz,
    acknowledged_at timestamptz, secret_deleted_at timestamptz,
    completed_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    onboarding platform.tenancy_onboarding_operations%ROWTYPE;
    delivery platform.tenancy_installation_credential_deliveries%ROWTYPE;
    now_at timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='installation delivery stage denied';
    END IF;
    IF p_attempt_id IS NULL OR p_credential_id IS NULL OR p_credential_version <> 1
       OR p_installation_manifest_sha256 !~ '^[0-9a-f]{64}$'
       OR p_delivery_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='installation delivery stage invalid';
    END IF;
    SELECT * INTO onboarding
    FROM platform.tenancy_onboarding_operations AS candidate
    WHERE candidate.operation_id=p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR onboarding.operation_kind <> 'environment_initialize'
       OR onboarding.state <> 'succeeded'
       OR onboarding.manifest_sha256 <> p_installation_manifest_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='installation delivery is not ready';
    END IF;
    SELECT * INTO delivery
    FROM platform.tenancy_installation_credential_deliveries AS candidate
    WHERE candidate.operation_id=p_operation_id
    FOR UPDATE;
    IF FOUND THEN
        IF delivery.tenant_id <> onboarding.tenant_id
           OR delivery.first_owner_user_id <> onboarding.first_user_id
           OR delivery.current_attempt_id <> p_attempt_id
           OR delivery.current_credential_id <> p_credential_id
           OR delivery.current_credential_version <> p_credential_version
           OR delivery.installation_manifest_sha256 <> p_installation_manifest_sha256
           OR delivery.delivery_manifest_sha256 <> p_delivery_manifest_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='installation delivery replay changed';
        END IF;
    ELSE
        INSERT INTO platform.tenancy_installation_credential_deliveries(
            operation_id,tenant_id,first_owner_user_id,installation_manifest_sha256,
            delivery_manifest_sha256,current_attempt_id,current_credential_id,
            current_credential_version,state,revision,created_at,updated_at)
        VALUES (p_operation_id,onboarding.tenant_id,onboarding.first_user_id,
            p_installation_manifest_sha256,p_delivery_manifest_sha256,p_attempt_id,
            p_credential_id,p_credential_version,'ready_to_display',1,now_at,now_at)
        RETURNING * INTO delivery;
        INSERT INTO platform.tenancy_installation_credential_delivery_attempts(
            attempt_id,operation_id,credential_id,credential_version,state,created_at)
        VALUES (p_attempt_id,p_operation_id,p_credential_id,p_credential_version,
            'ready_to_display',now_at);
    END IF;
    RETURN QUERY SELECT delivery.operation_id,delivery.tenant_id,
        delivery.first_owner_user_id,delivery.current_attempt_id,
        delivery.current_credential_id,delivery.current_credential_version,
        delivery.state,delivery.revision,now_at,delivery.claim_adapter_id,
        delivery.claim_lease_until,delivery.acknowledged_at,
        delivery.secret_deleted_at,delivery.completed_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_stage_v1(
    uuid,uuid,uuid,bigint,text,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_stage_v1(
    uuid,uuid,uuid,bigint,text,text) TO forgecrud_runtime;
