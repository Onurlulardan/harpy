-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_secret_deleted_v1(
    p_operation_id uuid, p_attempt_id uuid, p_expected_revision bigint
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
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='installation secret cleanup denied';
    END IF;
    SELECT * INTO delivery
    FROM platform.tenancy_installation_credential_deliveries AS candidate
    WHERE candidate.operation_id=p_operation_id AND candidate.current_attempt_id=p_attempt_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='installation delivery unavailable';
    END IF;
    IF delivery.state='secret_deleted' AND delivery.revision=p_expected_revision+1 THEN
        NULL;
    ELSIF delivery.state='acknowledged' AND delivery.revision=p_expected_revision THEN
        UPDATE platform.tenancy_installation_credential_deliveries AS target
        SET state='secret_deleted',secret_deleted_at=now_at,
            revision=target.revision+1,updated_at=now_at
        WHERE target.operation_id=p_operation_id
        RETURNING * INTO delivery;
    ELSE
        RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='installation cleanup state changed';
    END IF;
    RETURN QUERY SELECT delivery.state,delivery.revision,now_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_secret_deleted_v1(
    uuid,uuid,bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_secret_deleted_v1(
    uuid,uuid,bigint) TO forgecrud_runtime;
