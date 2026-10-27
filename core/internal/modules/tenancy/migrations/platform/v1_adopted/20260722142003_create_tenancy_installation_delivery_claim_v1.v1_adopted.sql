-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_claim_v1(
    p_operation_id uuid,
    p_attempt_id uuid,
    p_expected_revision bigint,
    p_adapter_id text,
    p_claim_token_sha256 text,
    p_lease_seconds integer
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
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='installation delivery claim denied';
    END IF;
    IF p_expected_revision < 1 OR p_attempt_id IS NULL
       OR octet_length(p_adapter_id) NOT BETWEEN 1 AND 128
       OR p_claim_token_sha256 !~ '^[0-9a-f]{64}$'
       OR p_lease_seconds NOT BETWEEN 10 AND 300 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='installation delivery claim invalid';
    END IF;
    SELECT * INTO delivery
    FROM platform.tenancy_installation_credential_deliveries AS candidate
    WHERE candidate.operation_id=p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR delivery.current_attempt_id <> p_attempt_id THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='installation delivery unavailable';
    END IF;
    IF delivery.state='claimed' AND delivery.revision=p_expected_revision+1
       AND delivery.claim_adapter_id=p_adapter_id
       AND delivery.claim_token_sha256=p_claim_token_sha256
       AND delivery.claim_lease_until>now_at THEN
        NULL;
    ELSIF delivery.state='ready_to_display' AND delivery.revision=p_expected_revision THEN
        UPDATE platform.tenancy_installation_credential_deliveries AS target
        SET state='claimed',claim_adapter_id=p_adapter_id,
            claim_token_sha256=p_claim_token_sha256,
            claim_lease_until=now_at+pg_catalog.make_interval(secs=>p_lease_seconds),
            claimed_at=now_at,revision=target.revision+1,updated_at=now_at
        WHERE target.operation_id=p_operation_id
        RETURNING * INTO delivery;
        UPDATE platform.tenancy_installation_credential_delivery_attempts AS attempt
        SET state='claimed',claimed_at=now_at
        WHERE attempt.attempt_id=p_attempt_id AND attempt.operation_id=p_operation_id
          AND attempt.state='ready_to_display';
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='installation delivery attempt changed';
        END IF;
    ELSE
        RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='installation delivery claim changed';
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

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_claim_v1(
    uuid,uuid,bigint,text,text,integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_claim_v1(
    uuid,uuid,bigint,text,text,integer) TO harpy_runtime;
