-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_forms_reference_source_v1(
    p_scope_schema name,p_component_id uuid
)
RETURNS TABLE(source_id uuid,source_version_id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog Forms reference source is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    WITH selected AS (
        SELECT candidate.projection_receipt_id
        FROM analytics_catalog_component_candidates AS candidate
        WHERE candidate.component_id=p_component_id
        UNION ALL
        SELECT head.projection_receipt_id
        FROM analytics_catalog_component_heads AS head
        WHERE head.component_id=p_component_id
          AND NOT EXISTS (
              SELECT 1
              FROM analytics_catalog_component_candidates AS candidate
              WHERE candidate.component_id=p_component_id
          )
    )
    SELECT projection.operational_resource_id,
           projection.published_version_id
    FROM selected
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=selected.projection_receipt_id
     AND projection.component_id=p_component_id
    JOIN analytics_catalog_components AS component
      ON component.component_id=projection.component_id
    WHERE projection.analytics_kind='endpoint'
      AND projection.operational_state='ready'
      AND projection.ready_at IS NOT NULL
      AND pg_catalog.uuid_extract_version(
          projection.operational_resource_id
      )=7
      AND pg_catalog.uuid_extract_version(projection.published_version_id)=7
      AND component.writer_state='v2'
      AND component.lifecycle_state IN ('AVAILABLE','STAGED','ACTIVE','DISABLED');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_forms_reference_source_v1(
    name,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_catalog_forms_reference_source_v1(
    name,uuid
) TO harpy_runtime;
