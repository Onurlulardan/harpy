-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_revision_json_v1(p_revision_id uuid)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    SELECT pg_catalog.jsonb_build_object(
        'id', revision.id,
        'resource_kind', revision.resource_kind,
        'version', revision.version,
        'state', revision.state,
        'definition', revision.definition_json::jsonb,
        'definition_sha256', pg_catalog.encode(
            revision.definition_sha256, 'hex'
        ),
        'created_at', revision.created_at,
        'updated_at', revision.updated_at,
        'published_at', revision.published_at,
        'row_revision', revision.row_revision
    )
    INTO v_document
    FROM system_resource_layout_revisions AS revision
    WHERE revision.id = p_revision_id;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_revision_json_v1(uuid)
    FROM PUBLIC;

