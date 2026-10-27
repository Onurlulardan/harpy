-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_profile_v1(
    p_schema_id uuid,
    p_field_id uuid,
    p_resource_kind text,
    p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_validation jsonb;
BEGIN
    v_validation := '{"maximum_bytes":128}'::jsonb;
    IF p_resource_kind NOT IN ('user', 'organization', 'role')
       OR p_now IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM installation profile input is invalid';
    END IF;
    INSERT INTO iam_profile_schema_versions (
        id, resource_kind, version, state, definition_sha256,
        activated_at, retired_at, created_at, updated_at, revision
    ) VALUES (
        p_schema_id, p_resource_kind, 1, 'draft',
        pg_catalog.sha256(pg_catalog.uuid_send(p_schema_id)),
        NULL, NULL, p_now, p_now, 1
    );
    INSERT INTO iam_profile_fields (
        id, resource_kind, field_key, value_kind, cardinality, required,
        classification, normalization_profile, state, created_at,
        updated_at, retired_at, revision
    ) VALUES (
        p_field_id, p_resource_kind, 'display_name', 'string', 'single', true,
        'public', 'none_v1', 'active', p_now, p_now, NULL, 1
    );
    INSERT INTO iam_profile_field_revisions (
        schema_version_id, resource_kind, field_id, validation_schema,
        validation_sha256, created_at, field_key, value_kind, cardinality,
        required, classification, normalization_profile, field_state,
        semantic_sha256
    ) VALUES (
        p_schema_id, p_resource_kind, p_field_id, v_validation,
        pg_catalog.sha256(pg_catalog.convert_to(
            iam_profile_validation_canonical_v1(v_validation), 'UTF8'
        )), p_now, 'display_name', 'string', 'single', true, 'public',
        'none_v1', 'active', iam_profile_semantic_sha256_v1(
            p_field_id, p_resource_kind, 'display_name', 'string',
            'single', true, 'public', 'none_v1'
        )
    );
    UPDATE iam_profile_schema_versions
    SET state = 'active',
        definition_sha256 = iam_profile_definition_sha256_v1(p_schema_id),
        activated_at = p_now, updated_at = p_now, revision = 2
    WHERE id = p_schema_id AND resource_kind = p_resource_kind
      AND state = 'draft' AND revision = 1;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM installation profile activation conflict';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_profile_v1(
    uuid, uuid, text, timestamptz
) FROM PUBLIC;
