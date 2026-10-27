-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_record_v1(
    p_operation_id uuid, p_current_attempt_id uuid, p_current_credential_id uuid,
    p_current_credential_version bigint, p_next_attempt_id uuid,
    p_next_credential_id uuid, p_next_credential_version bigint,
    p_expected_revision bigint, p_recovery_token_sha256 text,
    p_delivery_manifest_sha256 text, p_cause text
)
RETURNS TABLE (delivery_state text, delivery_revision bigint, database_now timestamptz)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    delivery platform.tenancy_installation_credential_deliveries%ROWTYPE;
    now_at timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='installation delivery record denied';
    END IF;
    IF p_expected_revision<1 OR p_cause NOT IN ('claim_lease_expired','outcome_unknown')
       OR p_delivery_manifest_sha256 !~ '^[0-9a-f]{64}$'
       OR p_recovery_token_sha256 !~ '^[0-9a-f]{64}$'
       OR p_next_credential_version<>p_current_credential_version+1 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='installation delivery record invalid';
    END IF;
    SELECT * INTO delivery
    FROM platform.tenancy_installation_credential_deliveries AS candidate
    WHERE candidate.operation_id=p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR delivery.state<>'outcome_unknown'
       OR delivery.revision<>p_expected_revision
       OR delivery.current_attempt_id<>p_current_attempt_id
       OR delivery.current_credential_id<>p_current_credential_id
       OR delivery.current_credential_version<>p_current_credential_version
       OR delivery.recovery_attempt_id<>p_next_attempt_id
       OR delivery.recovery_credential_id<>p_next_credential_id
       OR delivery.recovery_credential_version<>p_next_credential_version
       OR delivery.recovery_token_sha256<>p_recovery_token_sha256
       OR delivery.recovery_lease_until<=now_at OR delivery.recovery_cause<>p_cause
       OR delivery.delivery_manifest_sha256<>p_delivery_manifest_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='installation recovery binding changed';
    END IF;
    UPDATE platform.tenancy_installation_credential_delivery_attempts AS previous
    SET state='superseded',superseded_by_attempt_id=p_next_attempt_id,
        superseded_at=now_at
    WHERE previous.attempt_id=p_current_attempt_id
      AND previous.operation_id=p_operation_id
      AND previous.state='outcome_unknown' AND previous.cause=p_cause;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='installation recovery attempt changed';
    END IF;
    INSERT INTO platform.tenancy_installation_credential_delivery_attempts(
        attempt_id,operation_id,credential_id,credential_version,state,created_at)
    VALUES (p_next_attempt_id,p_operation_id,p_next_credential_id,
        p_next_credential_version,'ready_to_display',now_at);
    UPDATE platform.tenancy_installation_credential_deliveries AS target
    SET current_attempt_id=p_next_attempt_id,current_credential_id=p_next_credential_id,
        current_credential_version=p_next_credential_version,state='ready_to_display',
        claim_adapter_id=NULL,claim_token_sha256=NULL,claim_lease_until=NULL,
        claimed_at=NULL,acknowledged_at=NULL,secret_deleted_at=NULL,completed_at=NULL,
        recovery_attempt_id=NULL,recovery_credential_id=NULL,
        recovery_credential_version=NULL,recovery_token_sha256=NULL,
        recovery_lease_until=NULL,recovery_cause=NULL,
        revision=target.revision+1,updated_at=now_at
    WHERE target.operation_id=p_operation_id
    RETURNING * INTO delivery;
    RETURN QUERY SELECT delivery.state,delivery.revision,now_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_record_v1(
    uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint,text,text,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_record_v1(
    uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint,text,text,text)
    TO forgecrud_runtime;
