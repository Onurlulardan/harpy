-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_claim_v2(
    p_operation_id uuid,
    p_attempt_id uuid,
    p_lease_owner text,
    p_lease_token_sha256 text,
    p_expected_revision bigint,
    p_lease_seconds integer
)
RETURNS TABLE (
    operation_id uuid,
    attempt_id uuid,
    attempt_number bigint,
    revision bigint,
    lease_until timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    operation platform.tenancy_onboarding_operations%ROWTYPE;
    lease platform.tenancy_onboarding_leases%ROWTYPE;
    next_attempt bigint;
    next_state text;
    now_at timestamptz := pg_catalog.statement_timestamp();
    expires_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_lease_seconds < 1 OR p_lease_seconds > 300 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'onboarding claim denied';
    END IF;
    SELECT * INTO operation
    FROM platform.tenancy_onboarding_operations AS candidate
    WHERE candidate.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR operation.revision <> p_expected_revision
       OR operation.state NOT IN ('accepted', 'migrating', 'migration_ready',
            'authority_pending', 'authority_acknowledged') THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'onboarding claim revision conflict';
    END IF;
    SELECT * INTO lease
    FROM platform.tenancy_onboarding_leases AS current_lease
    WHERE current_lease.operation_id = p_operation_id
    FOR UPDATE;
    IF FOUND AND lease.lease_until > now_at THEN
        IF lease.attempt_id = p_attempt_id
           AND lease.lease_token_sha256 = p_lease_token_sha256
           AND lease.lease_owner = p_lease_owner THEN
            expires_at := now_at + pg_catalog.make_interval(secs => p_lease_seconds);
            UPDATE platform.tenancy_onboarding_leases
            SET lease_until = expires_at, heartbeat_at = now_at,
                revision = operation.revision
            WHERE tenancy_onboarding_leases.operation_id = p_operation_id;
            SELECT attempt.attempt_number INTO next_attempt
            FROM platform.tenancy_onboarding_attempts AS attempt
            WHERE attempt.attempt_id = lease.attempt_id;
            RETURN QUERY SELECT p_operation_id, p_attempt_id, next_attempt,
                operation.revision, expires_at;
            RETURN;
        END IF;
        RAISE EXCEPTION USING ERRCODE = '55P03',
            MESSAGE = 'onboarding lease is held';
    END IF;
    IF FOUND THEN
        UPDATE platform.tenancy_onboarding_attempts AS expired_attempt
        SET outcome = 'expired', finished_at = now_at
        WHERE expired_attempt.attempt_id = lease.attempt_id
          AND expired_attempt.outcome = 'active';
        DELETE FROM platform.tenancy_onboarding_leases
        WHERE tenancy_onboarding_leases.operation_id = p_operation_id;
    END IF;
    SELECT COALESCE(MAX(attempt.attempt_number), 0) + 1 INTO next_attempt
    FROM platform.tenancy_onboarding_attempts AS attempt
    WHERE attempt.operation_id = p_operation_id;
    expires_at := now_at + pg_catalog.make_interval(secs => p_lease_seconds);
    INSERT INTO platform.tenancy_onboarding_attempts (
        attempt_id, operation_id, attempt_number, outcome, started_at
    ) VALUES (p_attempt_id, p_operation_id, next_attempt, 'active', now_at);
    INSERT INTO platform.tenancy_onboarding_leases (
        operation_id, attempt_id, lease_owner, lease_token_sha256,
        lease_until, heartbeat_at, revision
    ) VALUES (p_operation_id, p_attempt_id, p_lease_owner,
        p_lease_token_sha256, expires_at, now_at, operation.revision + 1);
    next_state := CASE operation.state WHEN 'accepted' THEN 'migrating'
        ELSE operation.state END;
    UPDATE platform.tenancy_onboarding_operations
    SET state = next_state, revision = operation.revision + 1, updated_at = now_at
    WHERE tenancy_onboarding_operations.operation_id = p_operation_id;
    RETURN QUERY SELECT p_operation_id, p_attempt_id, next_attempt,
        operation.revision + 1, expires_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_claim_v2(
    uuid, uuid, text, text, bigint, integer
) FROM PUBLIC;
