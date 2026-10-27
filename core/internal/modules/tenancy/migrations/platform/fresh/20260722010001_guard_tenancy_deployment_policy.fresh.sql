-- +goose Up
CREATE TRIGGER tenancy_deployment_policy_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenancy_deployment_policy
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_get_deployment_policy_v1()
RETURNS TABLE (
    environment_id uuid,
    deployment_mode text,
    maximum_customer_tenants bigint,
    created_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
        OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING
            ERRCODE = '42501',
            MESSAGE = 'Tenancy deployment policy query denied';
    END IF;
    RETURN QUERY
    SELECT policy.environment_id,
           policy.deployment_mode,
           policy.maximum_customer_tenants,
           policy.created_at
    FROM platform.tenancy_deployment_policy AS policy
    WHERE policy.singleton
    FOR SHARE;
END;
$function$;
-- +goose StatementEnd
