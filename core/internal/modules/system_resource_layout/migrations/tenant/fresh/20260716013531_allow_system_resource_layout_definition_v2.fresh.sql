-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_definition_supported_v2(p_definition text)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN system_resource_layout_definition_valid_v1(p_definition)
        OR system_resource_layout_definition_valid_v2(p_definition);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_definition_supported_v2(text)
    FROM PUBLIC;

ALTER TABLE system_resource_layout_revisions
    DROP CONSTRAINT system_resource_layout_revisions_definition_check;

ALTER TABLE system_resource_layout_revisions
    ADD CONSTRAINT system_resource_layout_revisions_definition_check CHECK (
        system_resource_layout_definition_supported_v2(definition_json)
    );
