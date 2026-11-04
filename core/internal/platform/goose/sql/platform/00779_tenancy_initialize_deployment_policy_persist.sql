-- +goose Up
DROP FUNCTION platform.tenancy_initialize_deployment_policy_v1(uuid, uuid, text, text);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_initialize_deployment_policy_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_operation_kind text,
    in_deployment_mode text
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
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    WITH clock AS (
        SELECT pg_catalog.clock_timestamp() AS database_now
    ), inserted AS (
        INSERT INTO platform.tenancy_deployment_policy AS policy (
            environment_id, deployment_mode, maximum_customer_tenants, created_at
        )
        SELECT
            in_environment_id,
            in_deployment_mode,
            CASE in_deployment_mode
                WHEN 'onpremise' THEN 0::bigint
                WHEN 'saas' THEN NULL::bigint
            END,
            clock.database_now
        FROM clock
        WHERE in_operation_id IS NOT NULL
          AND in_operation_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND in_environment_id IS NOT NULL
          AND in_environment_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND in_deployment_mode IN ('onpremise', 'saas')
        ON CONFLICT (environment_id) DO NOTHING
        RETURNING
            policy.environment_id,
            policy.deployment_mode,
            policy.maximum_customer_tenants,
            policy.created_at,
            policy.created_at AS database_now
    )
    SELECT
        in_operation_id,
        inserted.environment_id,
        inserted.deployment_mode,
        inserted.maximum_customer_tenants,
        inserted.created_at,
        inserted.database_now,
        false
    FROM inserted
    UNION ALL
    SELECT
        in_operation_id,
        policy.environment_id,
        policy.deployment_mode,
        policy.maximum_customer_tenants,
        policy.created_at,
        clock.database_now,
        true
    FROM platform.tenancy_deployment_policy AS policy
    CROSS JOIN clock
    WHERE NOT EXISTS (SELECT 1 FROM inserted)
      AND policy.environment_id = in_environment_id
      AND policy.deployment_mode = in_deployment_mode
      AND in_operation_id IS NOT NULL
      AND in_operation_id <> '00000000-0000-0000-0000-000000000000'::uuid
      AND in_environment_id IS NOT NULL
      AND in_environment_id <> '00000000-0000-0000-0000-000000000000'::uuid
      AND in_deployment_mode IN ('onpremise', 'saas');
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_initialize_deployment_policy_v1(uuid, uuid, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_initialize_deployment_policy_v1(uuid, uuid, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_initialize_deployment_policy_v1(uuid, uuid, text, text);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_initialize_deployment_policy_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_operation_kind text,
    in_deployment_mode text
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
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_operation_id,
        in_environment_id,
        in_deployment_mode,
        NULL::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_initialize_deployment_policy_v1(uuid, uuid, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_initialize_deployment_policy_v1(uuid, uuid, text, text) TO harpy_api, harpy_worker;
