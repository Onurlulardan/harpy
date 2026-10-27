-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_ack_v1(
    p_operation_id uuid,
    p_attempt_id uuid,
    p_expected_revision bigint,
    p_adapter_id text,
    p_claim_token_sha256 text
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
    delivery platform.tenancy_installation_credential_deliveries%ROWTYPE;
    now_at timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='installation delivery ack denied';
    END IF;
    IF p_expected_revision < 1 OR p_attempt_id IS NULL
       OR octet_length(p_adapter_id) NOT BETWEEN 1 AND 128
       OR p_claim_token_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='installation delivery ack invalid';
    END IF;
    SELECT * INTO delivery
    FROM platform.tenancy_installation_credential_deliveries AS candidate
    WHERE candidate.operation_id=p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR delivery.current_attempt_id<>p_attempt_id
       OR delivery.claim_adapter_id<>p_adapter_id
       OR delivery.claim_token_sha256<>p_claim_token_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='installation delivery ack unavailable';
    END IF;
    IF delivery.state='acknowledged' AND delivery.revision=p_expected_revision+1 THEN
        NULL;
    ELSIF delivery.state='claimed' AND delivery.revision=p_expected_revision
       AND delivery.claim_lease_until>now_at THEN
        UPDATE platform.tenancy_installation_credential_deliveries AS target
        SET state='acknowledged',acknowledged_at=now_at,
            revision=target.revision+1,updated_at=now_at
        WHERE target.operation_id=p_operation_id
        RETURNING * INTO delivery;
        UPDATE platform.tenancy_installation_credential_delivery_attempts AS attempt
        SET state='acknowledged',acknowledged_at=now_at
        WHERE attempt.attempt_id=p_attempt_id AND attempt.operation_id=p_operation_id
          AND attempt.state='claimed';
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='installation delivery attempt changed';
        END IF;
    ELSE
        RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='installation delivery ack changed';
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

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_ack_v1(
    uuid,uuid,bigint,text,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_ack_v1(
    uuid,uuid,bigint,text,text) TO forgecrud_runtime;
