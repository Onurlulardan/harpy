-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_recover_v1(
    p_operation_id uuid, p_current_attempt_id uuid, p_current_credential_id uuid,
    p_current_credential_version bigint, p_next_attempt_id uuid,
    p_next_credential_id uuid, p_next_credential_version bigint,
    p_expected_revision bigint, p_recovery_token_sha256 text,
    p_delivery_manifest_sha256 text, p_lease_seconds integer
)
RETURNS TABLE (
    operation_id uuid, tenant_id uuid, first_owner_user_id uuid,
    current_credential_id uuid, current_credential_version bigint,
    next_credential_id uuid, next_credential_version bigint,
    current_attempt_id uuid, next_attempt_id uuid, recovery_cause text,
    delivery_manifest_sha256 text, delivery_revision bigint,
    database_now timestamptz, recovery_lease_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    delivery platform.tenancy_installation_credential_deliveries%ROWTYPE;
    attempt platform.tenancy_installation_credential_delivery_attempts%ROWTYPE;
    now_at timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='installation delivery recovery denied';
    END IF;
    IF p_expected_revision<1 OR p_next_attempt_id IS NULL OR p_next_credential_id IS NULL
       OR p_next_attempt_id=p_current_attempt_id OR p_next_credential_id=p_current_credential_id
       OR p_next_credential_version<>p_current_credential_version+1
       OR p_recovery_token_sha256 !~ '^[0-9a-f]{64}$'
       OR p_delivery_manifest_sha256 !~ '^[0-9a-f]{64}$'
       OR p_lease_seconds NOT BETWEEN 10 AND 300 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='installation delivery recovery invalid';
    END IF;
    SELECT * INTO delivery
    FROM platform.tenancy_installation_credential_deliveries AS candidate
    WHERE candidate.operation_id=p_operation_id
    FOR UPDATE;
    SELECT * INTO attempt
    FROM platform.tenancy_installation_credential_delivery_attempts AS candidate
    WHERE candidate.operation_id=p_operation_id AND candidate.attempt_id=p_current_attempt_id
    FOR UPDATE;
    IF delivery.operation_id IS NULL OR attempt.attempt_id IS NULL
       OR delivery.state<>'outcome_unknown' OR attempt.state<>'outcome_unknown'
       OR delivery.current_attempt_id<>p_current_attempt_id
       OR delivery.current_credential_id<>p_current_credential_id
       OR delivery.current_credential_version<>p_current_credential_version
       OR delivery.delivery_manifest_sha256<>p_delivery_manifest_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='installation delivery recovery unavailable';
    END IF;
    IF delivery.revision=p_expected_revision+1
       AND delivery.recovery_attempt_id=p_next_attempt_id
       AND delivery.recovery_credential_id=p_next_credential_id
       AND delivery.recovery_credential_version=p_next_credential_version
       AND delivery.recovery_token_sha256=p_recovery_token_sha256
       AND delivery.recovery_lease_until>now_at THEN
        NULL;
    ELSIF delivery.revision=p_expected_revision
       AND (delivery.recovery_lease_until IS NULL OR delivery.recovery_lease_until<=now_at) THEN
        UPDATE platform.tenancy_installation_credential_deliveries AS target
        SET recovery_attempt_id=p_next_attempt_id,
            recovery_credential_id=p_next_credential_id,
            recovery_credential_version=p_next_credential_version,
            recovery_token_sha256=p_recovery_token_sha256,
            recovery_lease_until=now_at+pg_catalog.make_interval(secs=>p_lease_seconds),
            recovery_cause=attempt.cause,revision=target.revision+1,updated_at=now_at
        WHERE target.operation_id=p_operation_id
        RETURNING * INTO delivery;
    ELSE
        RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='installation recovery lease changed';
    END IF;
    RETURN QUERY SELECT delivery.operation_id,delivery.tenant_id,
        delivery.first_owner_user_id,delivery.current_credential_id,
        delivery.current_credential_version,delivery.recovery_credential_id,
        delivery.recovery_credential_version,delivery.current_attempt_id,
        delivery.recovery_attempt_id,delivery.recovery_cause,
        delivery.delivery_manifest_sha256,delivery.revision,now_at,
        delivery.recovery_lease_until;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_recover_v1(
    uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint,text,text,integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_recover_v1(
    uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint,text,text,integer)
    TO forgecrud_runtime;
