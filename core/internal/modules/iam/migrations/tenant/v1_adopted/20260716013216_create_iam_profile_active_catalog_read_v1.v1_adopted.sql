-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_active_catalog_read_v1(
    p_scope_schema name,
    p_resource_kind text
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
            MESSAGE = 'IAM active profile catalog read scope is invalid';
    END IF;
    SELECT pg_catalog.jsonb_build_object(
        'schema_id', schema_version.id,
        'resource_kind', schema_version.resource_kind,
        'version', schema_version.version,
        'state', schema_version.state,
        'definition_sha256', pg_catalog.encode(
            schema_version.definition_sha256, 'hex'
        ),
        'fields', COALESCE(
            pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
                'id', snapshot.field_id,
                'key', snapshot.field_key,
                'logical_type', snapshot.value_kind,
                'cardinality', snapshot.cardinality,
                'required', snapshot.required,
                'classification', snapshot.classification,
                'normalization_profile', snapshot.normalization_profile,
                'state', snapshot.field_state,
                'validation', snapshot.validation_schema,
                'validation_sha256', pg_catalog.encode(
                    snapshot.validation_sha256, 'hex'
                ),
                'semantic_sha256', pg_catalog.encode(
                    snapshot.semantic_sha256, 'hex'
                )
            ) ORDER BY snapshot.field_id) FILTER (
                WHERE snapshot.field_id IS NOT NULL
            ),
            '[]'::jsonb
        )
    )
    INTO v_document
    FROM iam_profile_schema_versions AS schema_version
    LEFT JOIN iam_profile_field_revisions AS snapshot
      ON snapshot.schema_version_id = schema_version.id
     AND snapshot.resource_kind = schema_version.resource_kind
    WHERE schema_version.resource_kind = p_resource_kind
      AND schema_version.state = 'active'
    GROUP BY schema_version.id, schema_version.resource_kind,
             schema_version.version, schema_version.state,
             schema_version.definition_sha256;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_active_catalog_read_v1(name,text)
    FROM PUBLIC;
