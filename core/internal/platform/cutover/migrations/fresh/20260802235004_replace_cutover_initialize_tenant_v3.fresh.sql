-- +goose Up
-- Migration stream version: 20260802235004.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.cutover_initialize_tenant_v3(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_owner_user_id uuid,
    p_manifest_sha256 text,
    p_operation_kind text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    PERFORM platform.cutover_initialize_tenant_v4(
        p_operation_id, p_tenant_id, p_owner_user_id,
        p_manifest_sha256, p_operation_kind
    );
END;
$function$;
-- +goose StatementEnd
