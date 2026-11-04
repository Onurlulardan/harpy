-- +goose Up
CREATE FUNCTION platform.tenancy_authority_reconcile_current_v1(
    in_operation_id uuid,
    in_manifest_sha256 text
)
RETURNS TABLE (
    billing_global_reconciled integer,
    billing_global_remaining integer,
    billing_tenant_reconciled integer,
    billing_tenant_remaining integer,
    analytics_reconciled integer,
    analytics_remaining integer
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        0::integer,
        0::integer,
        0::integer,
        0::integer,
        0::integer,
        0::integer;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_authority_reconcile_current_v1(uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_authority_reconcile_current_v1(uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_authority_reconcile_current_v1(uuid, text);
