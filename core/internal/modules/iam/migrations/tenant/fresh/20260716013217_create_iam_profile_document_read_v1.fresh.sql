-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_document_read_v1(
    p_scope_schema name,
    p_resource_kind text,
    p_record_id uuid
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_resource_kind NOT IN ('user', 'organization', 'role') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile document read scope is invalid';
    END IF;
    IF p_resource_kind = 'user' THEN
        SELECT pg_catalog.jsonb_build_object(
            'schema_id', profile_schema_version_id,
            'values', profile_values,
            'sha256', pg_catalog.encode(profile_sha256, 'hex')
        )
        INTO v_document
        FROM iam_users WHERE id = p_record_id;
    ELSIF p_resource_kind = 'organization' THEN
        SELECT pg_catalog.jsonb_build_object(
            'schema_id', profile_schema_version_id,
            'values', profile_values,
            'sha256', pg_catalog.encode(profile_sha256, 'hex')
        )
        INTO v_document
        FROM iam_organizations WHERE id = p_record_id;
    ELSE
        SELECT pg_catalog.jsonb_build_object(
            'schema_id', profile_schema_version_id,
            'values', profile_values,
            'sha256', pg_catalog.encode(profile_sha256, 'hex')
        )
        INTO v_document
        FROM iam_roles WHERE id = p_record_id;
    END IF;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_document_read_v1(name,text,uuid)
    FROM PUBLIC;
