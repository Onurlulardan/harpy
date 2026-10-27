-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_owner_purge_authority_read_v2(
    p_scope_schema name,p_operation_id uuid,p_component_id uuid,
    p_package_id uuid,p_generation bigint,p_contract text,
    p_content_sha256 text
)
RETURNS SETOF analytics_catalog_owner_purge_authorities_v2
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_existing analytics_catalog_owner_purge_authorities_v2%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_package_id) IS DISTINCT FROM 7 OR
       p_generation<1 OR p_contract IS NULL OR
       p_content_sha256!~'^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog purge authority read is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT authority.* INTO v_existing
    FROM analytics_catalog_owner_purge_authorities_v2 AS authority
    JOIN analytics_catalog_components AS component
      ON component.component_id=authority.component_id
     AND component.owner_generation=p_generation
     AND component.contract=p_contract
     AND component.content_sha256=p_content_sha256
     AND component.writer_state='v2'
     AND component.lifecycle_state IN ('RETIRED','PURGED')
    JOIN analytics_catalog_version_projections AS projection
      ON projection.component_id=component.component_id
     AND projection.package_id=p_package_id
     AND projection.operational_resource_id=authority.resource_id
     AND projection.published_version_id=authority.version_id
     AND projection.receipt_id=authority.head_id
    WHERE authority.operation_id=p_operation_id
      AND authority.component_id=p_component_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge authority is unavailable';
    END IF;
    RETURN NEXT v_existing;
END;
$function$;

CREATE FUNCTION analytics_catalog_owner_purge_authority_guard_v2()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics Catalog purge authority is immutable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_owner_purge_authority_read_v2(
    name,uuid,uuid,uuid,bigint,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_catalog_owner_purge_authority_read_v2(
    name,uuid,uuid,uuid,bigint,text,text
) TO harpy_runtime;
REVOKE ALL ON FUNCTION
    analytics_catalog_owner_purge_authority_guard_v2() FROM PUBLIC;

CREATE TRIGGER analytics_catalog_owner_purge_authority_immutable_v2
BEFORE UPDATE OR DELETE ON analytics_catalog_owner_purge_authorities_v2
FOR EACH ROW EXECUTE FUNCTION
    analytics_catalog_owner_purge_authority_guard_v2();
ALTER TABLE analytics_catalog_owner_purge_authorities_v2
    ENABLE ALWAYS TRIGGER
    analytics_catalog_owner_purge_authority_immutable_v2;
