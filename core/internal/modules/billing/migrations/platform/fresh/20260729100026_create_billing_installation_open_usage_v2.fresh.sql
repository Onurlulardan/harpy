-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_open_usage_v2(
    p_operation_id uuid, p_tenant_id uuid, p_manifest_sha256 text
)
RETURNS TABLE (meter_count integer, replay boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR platform.tenancy_activation_step_consume_v1(
           p_operation_id, 'billing.open_tenant_usage'
       ) IS DISTINCT FROM TRUE THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing installation usage opening is denied';
    END IF;
    RETURN QUERY
    SELECT opening.meter_count, opening.replay
    FROM platform.billing_installation_open_usage_v1(
        p_operation_id, p_tenant_id, p_manifest_sha256
    ) AS opening;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_installation_open_usage_v2(
    uuid,uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_installation_open_usage_v2(
    uuid,uuid,text
) TO forgecrud_runtime;
