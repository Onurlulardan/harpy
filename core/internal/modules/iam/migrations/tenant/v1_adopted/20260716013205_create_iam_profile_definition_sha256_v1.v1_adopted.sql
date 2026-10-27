-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_definition_sha256_v1(p_schema_id uuid)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource_kind text;
    v_version bigint;
    v_fields text;
    v_document text;
BEGIN
    SELECT schema_version.resource_kind, schema_version.version
    INTO v_resource_kind, v_version
    FROM iam_profile_schema_versions AS schema_version
    WHERE schema_version.id = p_schema_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM profile schema is unknown';
    END IF;
    SELECT COALESCE(
        '[' || pg_catalog.string_agg(
            '{"cardinality":' ||
                pg_catalog.to_jsonb(snapshot.cardinality)::text ||
            ',"classification":' ||
                pg_catalog.to_jsonb(snapshot.classification)::text ||
            ',"id":' || pg_catalog.to_jsonb(snapshot.field_id::text)::text ||
            ',"key":' || pg_catalog.to_jsonb(snapshot.field_key)::text ||
            ',"logical_type":' ||
                pg_catalog.to_jsonb(snapshot.value_kind)::text ||
            ',"normalization_profile":' ||
                pg_catalog.to_jsonb(snapshot.normalization_profile)::text ||
            ',"required":' || pg_catalog.to_jsonb(snapshot.required)::text ||
            ',"state":' || pg_catalog.to_jsonb(snapshot.field_state)::text ||
            ',"validation":' ||
                iam_profile_validation_canonical_v1(
                    snapshot.validation_schema
                ) || '}',
            ',' ORDER BY snapshot.field_id::text
        ) || ']',
        '[]'
    )
    INTO v_fields
    FROM iam_profile_field_revisions AS snapshot
    WHERE snapshot.schema_version_id = p_schema_id;
    v_document := '{"contract":"harpy/iam-profile-schema/v1"' ||
        ',"fields":' || v_fields ||
        ',"id":' || pg_catalog.to_jsonb(p_schema_id::text)::text ||
        ',"resource_kind":' || pg_catalog.to_jsonb(v_resource_kind)::text ||
        ',"version":' || v_version::text || '}';
    RETURN pg_catalog.sha256(pg_catalog.convert_to(v_document, 'UTF8'));
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_definition_sha256_v1(uuid) FROM PUBLIC;
