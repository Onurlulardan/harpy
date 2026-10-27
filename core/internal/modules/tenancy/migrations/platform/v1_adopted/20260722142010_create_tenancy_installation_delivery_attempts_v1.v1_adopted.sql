-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_attempts_v1(p_operation_id uuid)
RETURNS TABLE (
    attempt_id uuid, credential_id uuid, credential_version bigint,
    attempt_state text, cause text, superseded_by_attempt_id uuid,
    created_at timestamptz, claimed_at timestamptz,
    acknowledged_at timestamptz, superseded_at timestamptz,
    completed_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='installation delivery attempts denied';
    END IF;
    RETURN QUERY SELECT attempt.attempt_id,attempt.credential_id,attempt.credential_version,
        attempt.state,attempt.cause,attempt.superseded_by_attempt_id,
        attempt.created_at,attempt.claimed_at,attempt.acknowledged_at,
        attempt.superseded_at,attempt.completed_at
    FROM platform.tenancy_installation_credential_delivery_attempts AS attempt
    WHERE attempt.operation_id=p_operation_id
    ORDER BY attempt.credential_version;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_attempts_v1(uuid)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_attempts_v1(uuid)
    TO harpy_runtime;
