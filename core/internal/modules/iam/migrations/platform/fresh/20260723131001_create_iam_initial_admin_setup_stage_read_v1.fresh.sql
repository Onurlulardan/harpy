-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_initial_admin_setup_stage_read_v1(
    p_operation_id uuid,
    p_tenant_id uuid
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_stage jsonb;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_operation_id IS NULL OR p_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='initial admin setup stage read is denied';
    END IF;
    SELECT pg_catalog.to_jsonb(stage) INTO STRICT v_stage
    FROM platform.iam_initial_admin_setup_stages AS stage
    WHERE stage.operation_id=p_operation_id
      AND stage.tenant_id=p_tenant_id
    FOR SHARE;
    RETURN v_stage;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_initial_admin_setup_stage_read_v1(
    uuid,uuid
) FROM PUBLIC;
