-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_submission_step_v2(
    p_plan_sha256 text, p_resource_kind text, p_document bytea,
    p_expected_ordinal integer, p_allowed_operations jsonb
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_step jsonb;
BEGIN
    v_step := platform.platform_migrator_validate_submission_step_v1(
        p_plan_sha256,p_resource_kind,p_document,p_expected_ordinal,
        p_allowed_operations
    );
    IF pg_catalog.jsonb_typeof(v_step#>'{identifier_mapping,form}')<>'object'
       OR v_step#>>'{identifier_mapping,form,logical_id}' IS NULL
       OR v_step#>>'{identifier_mapping,form,physical_name}' IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='schema V2 submission mapping differs';
    END IF;
    RETURN v_step;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_validate_submission_step_v2(
    text,text,bytea,integer,jsonb
) FROM PUBLIC;
