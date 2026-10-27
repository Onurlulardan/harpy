-- +goose Up
CREATE TABLE analytics_catalog_owner_purge_authorities_v2 (
    operation_id uuid NOT NULL,
    component_id uuid NOT NULL,
    reference_epoch bigint NOT NULL CHECK (reference_epoch>0),
    authority_sha256 text NOT NULL
        CHECK (authority_sha256~'^[0-9a-f]{64}$'),
    resource_id uuid NOT NULL,
    version_id uuid NOT NULL,
    head_id uuid NOT NULL,
    recorded_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id,component_id),
    FOREIGN KEY (component_id)
        REFERENCES analytics_catalog_components(component_id)
        ON DELETE RESTRICT,
    CHECK (
        pg_catalog.uuid_extract_version(operation_id)=7 AND
        pg_catalog.uuid_extract_version(component_id)=7 AND
        pg_catalog.uuid_extract_version(resource_id)=7 AND
        pg_catalog.uuid_extract_version(version_id)=7 AND
        pg_catalog.uuid_extract_version(head_id)=7 AND
        pg_catalog.isfinite(recorded_at) AND
        date_trunc('microseconds',recorded_at)=recorded_at
    )
);
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_owner_binding_v2(
    p_scope_schema name,p_component_id uuid,p_package_id uuid,
    p_generation bigint,p_contract text,p_content_sha256 text
)
RETURNS TABLE(
    resource_id uuid,version_id uuid,head_id uuid,
    lifecycle_state text,resource_reference_epoch bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_count bigint;
BEGIN
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_package_id) IS DISTINCT FROM 7 OR
       p_generation<1 OR p_contract IS NULL OR
       p_content_sha256!~'^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog owner binding input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    SELECT projection.operational_resource_id,
           projection.published_version_id,
           head.projection_receipt_id,component.lifecycle_state,
           root.reference_epoch
    FROM analytics_catalog_components AS component
    JOIN analytics_catalog_component_heads AS head
      ON head.component_id=component.component_id
     AND head.package_id=p_package_id
     AND head.content_sha256=component.content_sha256
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=head.projection_receipt_id
     AND projection.component_id=head.component_id
     AND projection.package_id=head.package_id
     AND projection.package_version=head.package_version
     AND projection.release_id=head.release_id
     AND projection.content_sha256=head.content_sha256
     AND projection.analytics_kind='endpoint'
     AND projection.operational_state='ready'
     AND projection.ready_at IS NOT NULL
    JOIN analytics_saved_queries AS root
      ON root.id=projection.operational_resource_id
    JOIN analytics_saved_query_versions AS version_row
      ON version_row.query_id=root.id
     AND version_row.id=projection.published_version_id
    WHERE component.component_id=p_component_id
      AND component.owner_generation=p_generation
      AND component.contract=p_contract
      AND component.content_sha256=p_content_sha256
      AND component.writer_state='v2'
    UNION ALL
    SELECT projection.operational_resource_id,
           projection.published_version_id,
           head.projection_receipt_id,component.lifecycle_state,
           root.reference_epoch
    FROM analytics_catalog_components AS component
    JOIN analytics_catalog_component_heads AS head
      ON head.component_id=component.component_id
     AND head.package_id=p_package_id
     AND head.content_sha256=component.content_sha256
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=head.projection_receipt_id
     AND projection.component_id=head.component_id
     AND projection.package_id=head.package_id
     AND projection.package_version=head.package_version
     AND projection.release_id=head.release_id
     AND projection.content_sha256=head.content_sha256
     AND projection.analytics_kind='dashboard'
     AND projection.operational_state='ready'
     AND projection.ready_at IS NOT NULL
    JOIN analytics_dashboards AS root
      ON root.id=projection.operational_resource_id
    JOIN analytics_dashboard_versions AS version_row
      ON version_row.dashboard_id=root.id
     AND version_row.id=projection.published_version_id
    WHERE component.component_id=p_component_id
      AND component.owner_generation=p_generation
      AND component.contract=p_contract
      AND component.content_sha256=p_content_sha256
      AND component.writer_state='v2';
    GET DIAGNOSTICS v_count=ROW_COUNT;
    IF v_count<>1 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog owner binding differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_owner_binding_v2(
    name,uuid,uuid,bigint,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_catalog_owner_binding_v2(
    name,uuid,uuid,bigint,text,text
) TO forgecrud_runtime;
