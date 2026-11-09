-- +goose Up
DROP FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_claim_v2(
    in_operation_id uuid, in_attempt_id uuid, in_lease_owner text, in_lease_token_sha256 text,
    in_expected_revision bigint, in_lease_seconds integer
) RETURNS TABLE (operation_id uuid, attempt_id uuid, attempt_number bigint, revision bigint, lease_until timestamptz)
LANGUAGE plpgsql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
DECLARE
    operation platform.tenancy_onboarding_operations%ROWTYPE;
    lease platform.tenancy_onboarding_leases%ROWTYPE;
    next_attempt bigint;
    until_at timestamptz;
BEGIN
    IF in_lease_seconds < 1 OR in_lease_seconds > 3600 OR in_lease_owner IS NULL OR in_lease_owner = ''
        OR in_lease_token_sha256 !~ '^[0-9a-f]{64}$' OR in_attempt_id IS NULL THEN
        RAISE EXCEPTION 'onboarding claim is invalid' USING ERRCODE = '22023';
    END IF;
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations AS stored
    WHERE stored.operation_id = in_operation_id FOR UPDATE;
    IF NOT FOUND OR operation.revision <> in_expected_revision THEN
        RAISE EXCEPTION 'onboarding claim conflict' USING ERRCODE = '40001';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.tenancy_onboarding_leases AS other
        WHERE other.lease_token_sha256 = in_lease_token_sha256 AND other.operation_id <> in_operation_id
          AND other.lease_until > pg_catalog.clock_timestamp()) THEN
        RAISE EXCEPTION 'onboarding lease belongs to another operation' USING ERRCODE = '42501';
    END IF;
    SELECT * INTO lease FROM platform.tenancy_onboarding_leases AS stored WHERE stored.operation_id = in_operation_id;
    IF FOUND AND lease.lease_until > pg_catalog.clock_timestamp() THEN
        IF lease.lease_token_sha256 = in_lease_token_sha256 AND lease.lease_owner = in_lease_owner
           AND lease.tenant_id = operation.tenant_id THEN
            RETURN QUERY SELECT operation.operation_id, lease.attempt_id, operation.attempt_number, operation.revision, lease.lease_until;
            RETURN;
        END IF;
        RAISE EXCEPTION 'onboarding lease is held' USING ERRCODE = '55P03';
    END IF;
    next_attempt := operation.attempt_number + 1;
    until_at := pg_catalog.clock_timestamp() + (in_lease_seconds * pg_catalog.interval '1 second');
    UPDATE platform.tenancy_onboarding_operations SET attempt_number = next_attempt, updated_at = pg_catalog.clock_timestamp()
    WHERE tenancy_onboarding_operations.operation_id = in_operation_id;
    INSERT INTO platform.tenancy_onboarding_attempts (
        attempt_id, operation_id, tenant_id, attempt_number, lease_owner, lease_token_sha256, created_at
    ) VALUES (
        in_attempt_id, in_operation_id, operation.tenant_id, next_attempt, in_lease_owner, in_lease_token_sha256, pg_catalog.clock_timestamp()
    );
    INSERT INTO platform.tenancy_onboarding_leases (
        operation_id, tenant_id, attempt_id, lease_owner, lease_token_sha256, lease_until
    ) VALUES (
        in_operation_id, operation.tenant_id, in_attempt_id, in_lease_owner, in_lease_token_sha256, until_at
    ) ON CONFLICT ON CONSTRAINT tenancy_onboarding_leases_pkey DO UPDATE SET
        tenant_id = EXCLUDED.tenant_id, attempt_id = EXCLUDED.attempt_id, lease_owner = EXCLUDED.lease_owner,
        lease_token_sha256 = EXCLUDED.lease_token_sha256, lease_until = EXCLUDED.lease_until;
    RETURN QUERY SELECT in_operation_id, in_attempt_id, next_attempt, operation.revision, until_at;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer);
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_claim_v2(
    in_operation_id uuid, in_attempt_id uuid, in_lease_owner text, in_lease_token_sha256 text,
    in_expected_revision bigint, in_lease_seconds integer
) RETURNS TABLE (operation_id uuid, attempt_id uuid, attempt_number bigint, revision bigint, lease_until timestamptz)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    UPDATE platform.tenancy_onboarding_operations AS operation
    SET attempt_number = operation.attempt_number + 1, updated_at = pg_catalog.clock_timestamp()
    WHERE operation.operation_id = in_operation_id AND operation.revision = in_expected_revision
    RETURNING operation.operation_id, in_attempt_id, operation.attempt_number, operation.revision,
        pg_catalog.clock_timestamp() + (in_lease_seconds * pg_catalog.interval '1 second');
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer) TO harpy_api, harpy_worker;
