-- +goose Up
-- Migration stream version: 20260722080102.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_cutover_initialize_v1(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_owner_user_id uuid,
    p_manifest_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant cutover activation denied';
    END IF;
    PERFORM platform.cutover_initialize_tenant_v1(
        p_operation_id, p_tenant_id, p_owner_user_id, p_manifest_sha256
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_activation_cutover_initialize_v1(
    uuid, uuid, uuid, text
) FROM PUBLIC;
