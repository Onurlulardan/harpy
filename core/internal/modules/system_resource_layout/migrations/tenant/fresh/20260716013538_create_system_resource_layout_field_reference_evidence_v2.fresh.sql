-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_field_reference_evidence_v2(
    p_scope_schema name,
    p_resource_kind text,
    p_field_id uuid
)
RETURNS TABLE (
    live_reference_count bigint,
    manifest_sha256 bytea,
    checked_owner_count integer
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_count bigint;
    v_manifest text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_resource_kind NOT IN ('user', 'organization', 'role')
       OR p_field_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'system resource layout reference scope is invalid';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(p_field_id::text, 8501)
    );
    SELECT pg_catalog.count(*)
    INTO v_count
    FROM system_resource_layout_publications AS publication
    JOIN system_resource_layout_revisions AS revision
      ON revision.id = publication.revision_id
    WHERE publication.resource_kind = p_resource_kind
      AND revision.state = 'published'
      AND system_resource_layout_field_ids_v2(revision.definition_json)
          ? p_field_id::text;
    v_manifest := '{"checked_owner_count":1' ||
        ',"contract":"harpy/system-resource-layout-reference/v1"' ||
        ',"field_id":' || pg_catalog.to_jsonb(p_field_id::text)::text ||
        ',"live_reference_count":' || v_count::text ||
        ',"resource_kind":' || pg_catalog.to_jsonb(p_resource_kind)::text || '}';
    RETURN QUERY SELECT v_count,
        pg_catalog.sha256(pg_catalog.convert_to(v_manifest, 'UTF8')), 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_field_reference_evidence_v2(
    name,text,uuid
) FROM PUBLIC;
