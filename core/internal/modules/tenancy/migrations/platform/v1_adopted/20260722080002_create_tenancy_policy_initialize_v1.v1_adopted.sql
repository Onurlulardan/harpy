-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_initialize_deployment_policy_v1(
    p_operation_id uuid,
    p_environment_id uuid,
    p_operation_kind text,
    p_deployment_mode text
)
RETURNS TABLE (
    operation_id uuid,
    environment_id uuid,
    deployment_mode text,
    maximum_customer_tenants bigint,
    created_at timestamptz,
    database_now timestamptz,
    replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    initialization platform.tenancy_deployment_policy_initializations%ROWTYPE;
    policy platform.tenancy_deployment_policy%ROWTYPE;
    capacity bigint;
    now_at timestamptz := pg_catalog.statement_timestamp();
    policy_preexisted boolean := false;
BEGIN
    capacity := CASE p_deployment_mode WHEN 'onpremise' THEN 0 ELSE NULL END;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_operation_kind IS DISTINCT FROM 'environment_initialize'
       OR p_deployment_mode NOT IN ('onpremise', 'saas') THEN
        RAISE EXCEPTION USING
            ERRCODE = '42501', MESSAGE = 'deployment policy initialization denied';
    END IF;
    PERFORM 1 FROM platform.environment_identity AS identity
    WHERE identity.environment_id = p_environment_id
    FOR NO KEY UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING
            ERRCODE = 'P0002', MESSAGE = 'environment identity is missing';
    END IF;
    SELECT * INTO initialization
    FROM platform.tenancy_deployment_policy_initializations AS candidate
    WHERE candidate.environment_id = p_environment_id
    FOR UPDATE;
    IF FOUND THEN
        IF initialization.operation_id IS DISTINCT FROM p_operation_id
           OR initialization.operation_kind IS DISTINCT FROM p_operation_kind
           OR initialization.deployment_mode IS DISTINCT FROM p_deployment_mode
           OR initialization.maximum_customer_tenants IS DISTINCT FROM capacity THEN
            RAISE EXCEPTION USING
                ERRCODE = '23505', MESSAGE = 'deployment policy initialization conflicts';
        END IF;
        SELECT * INTO policy
        FROM platform.tenancy_deployment_policy AS existing
        WHERE existing.environment_id = p_environment_id
        FOR UPDATE;
        IF NOT FOUND OR policy.deployment_mode IS DISTINCT FROM p_deployment_mode
           OR policy.maximum_customer_tenants IS DISTINCT FROM capacity THEN
            RAISE EXCEPTION USING
                ERRCODE = '23505', MESSAGE = 'deployment policy state conflicts';
        END IF;
        RETURN QUERY SELECT initialization.operation_id, policy.environment_id,
            policy.deployment_mode, policy.maximum_customer_tenants,
            policy.created_at, now_at, true;
        RETURN;
    END IF;
    SELECT * INTO policy
    FROM platform.tenancy_deployment_policy AS existing
    WHERE existing.singleton
    FOR UPDATE;
    IF FOUND THEN
        policy_preexisted := true;
        IF policy.environment_id IS DISTINCT FROM p_environment_id
           OR policy.deployment_mode IS DISTINCT FROM p_deployment_mode
           OR policy.maximum_customer_tenants IS DISTINCT FROM capacity THEN
            RAISE EXCEPTION USING
                ERRCODE = '23505', MESSAGE = 'deployment policy state conflicts';
        END IF;
    ELSE
        INSERT INTO platform.tenancy_deployment_policy (
            environment_id, deployment_mode, maximum_customer_tenants, created_at
        ) VALUES (p_environment_id, p_deployment_mode, capacity, now_at)
        RETURNING * INTO policy;
    END IF;
    INSERT INTO platform.tenancy_deployment_policy_initializations (
        operation_id, environment_id, operation_kind, deployment_mode,
        maximum_customer_tenants, created_at
    ) VALUES (p_operation_id, p_environment_id, p_operation_kind,
        p_deployment_mode, capacity, now_at);
    RETURN QUERY SELECT p_operation_id, policy.environment_id,
        policy.deployment_mode, policy.maximum_customer_tenants,
        policy.created_at, now_at, policy_preexisted;
END;
$function$;
-- +goose StatementEnd
