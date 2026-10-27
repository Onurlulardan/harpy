-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_write_resource_read_v1(
    p_scope_schema name, p_resource_kind text, p_resource_id uuid
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_resource_kind NOT IN ('user','organization','role') OR
       p_resource_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile write resource scope is invalid';
    END IF;
    IF p_resource_kind = 'user' THEN
        SELECT pg_catalog.jsonb_build_object(
            'schema_id', resource.profile_schema_version_id,
            'values', resource.profile_values,
            'sha256', pg_catalog.encode(resource.profile_sha256, 'hex'),
            'revision', resource.revision
        ) INTO v_document
        FROM iam_users AS resource WHERE resource.id = p_resource_id
        FOR UPDATE;
    ELSIF p_resource_kind = 'organization' THEN
        SELECT pg_catalog.jsonb_build_object(
            'schema_id', resource.profile_schema_version_id,
            'values', resource.profile_values,
            'sha256', pg_catalog.encode(resource.profile_sha256, 'hex'),
            'revision', resource.revision
        ) INTO v_document
        FROM iam_organizations AS resource WHERE resource.id = p_resource_id
        FOR UPDATE;
    ELSE
        SELECT pg_catalog.jsonb_build_object(
            'schema_id', resource.profile_schema_version_id,
            'values', resource.profile_values,
            'sha256', pg_catalog.encode(resource.profile_sha256, 'hex'),
            'revision', resource.revision
        ) INTO v_document
        FROM iam_roles AS resource WHERE resource.id = p_resource_id
        FOR UPDATE;
    END IF;
    IF v_document IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM profile write resource is unavailable';
    END IF;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_write_resource_read_v1(name,text,uuid)
    FROM PUBLIC;
