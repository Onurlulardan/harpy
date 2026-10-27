-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_profile_assert_v1(
    p_resource_kind text, p_schema_id uuid,
    p_profile_document bytea, p_profile_sha256 bytea
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_values jsonb;
BEGIN
    IF p_resource_kind NOT IN ('user', 'organization', 'role') OR
       p_schema_id IS NULL OR pg_catalog.octet_length(p_profile_document) > 1048576 OR
       pg_catalog.octet_length(p_profile_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.sha256(p_profile_document) IS DISTINCT FROM p_profile_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM identity profile input is invalid';
    END IF;
    v_values := pg_catalog.convert_from(p_profile_document, 'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_values) <> 'object' OR NOT EXISTS (
        SELECT 1 FROM iam_profile_schema_versions AS schema_version
        WHERE schema_version.id = p_schema_id
          AND schema_version.resource_kind = p_resource_kind
          AND schema_version.state = 'active'
    ) OR EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_each(v_values) AS entry(key, value)
        WHERE entry.key !~
            '^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
           OR pg_catalog.jsonb_typeof(entry.value) = 'null'
           OR NOT EXISTS (
              SELECT 1 FROM iam_profile_field_revisions AS field
              WHERE field.schema_version_id = p_schema_id
                AND field.resource_kind = p_resource_kind
                AND field.field_id = entry.key::uuid AND field.field_state = 'active'
           )
    ) OR EXISTS (
        SELECT 1 FROM iam_profile_field_revisions AS required
        WHERE required.schema_version_id = p_schema_id
          AND required.resource_kind = p_resource_kind
          AND required.field_state = 'active' AND required.required
          AND NOT v_values ? required.field_id::text
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM identity profile contract is invalid';
    END IF;
    RETURN v_values;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_profile_assert_v1(
    text,uuid,bytea,bytea
) FROM PUBLIC;
