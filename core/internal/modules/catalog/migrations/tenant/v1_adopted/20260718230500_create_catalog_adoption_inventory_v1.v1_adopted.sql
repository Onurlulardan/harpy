-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_adoption_inventory_v1(p_tenant_id uuid)
RETURNS TABLE(side text, kind text, identity text, document text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER NOT IN ('harpy_v2_runtime','harpy_adopter') OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR NOT EXISTS (
        SELECT 1 FROM tenancy_schema_identity AS tenant
        WHERE tenant.singleton AND tenant.tenant_id=p_tenant_id
          AND tenant.schema_name=pg_catalog.current_schema()
    ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog adoption inventory authority denied';
    END IF;
    IF (pg_catalog.to_regclass('modules') IS NULL) IS DISTINCT FROM
       (pg_catalog.to_regclass('module_dependencies') IS NULL) OR
       (pg_catalog.to_regclass('modules') IS NULL) IS DISTINCT FROM
       (pg_catalog.to_regclass('module_forms') IS NULL) OR
       (pg_catalog.to_regclass('modules') IS NULL) IS DISTINCT FROM
       (pg_catalog.to_regclass('module_schema_versions') IS NULL) OR
       (pg_catalog.to_regclass('modules') IS NULL) IS DISTINCT FROM
       (pg_catalog.to_regclass('module_migration_runs') IS NULL) OR
       (pg_catalog.to_regclass('modules') IS NULL) IS DISTINCT FROM
       (pg_catalog.to_regclass('menu_configs') IS NULL) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog adoption legacy inventory is incomplete';
    END IF;
    IF pg_catalog.to_regclass('modules') IS NOT NULL THEN
        RETURN QUERY SELECT 'source','module',row.id::text,to_jsonb(row)::text
        FROM modules AS row WHERE row.tenant_id=p_tenant_id
          AND row.deleted_at IS NULL;
    END IF;
    IF pg_catalog.to_regclass('module_dependencies') IS NOT NULL THEN
        RETURN QUERY SELECT 'source','dependency',row.id::text,to_jsonb(row)::text
        FROM module_dependencies AS row JOIN modules AS owner
          ON owner.id=row.module_id
        WHERE owner.tenant_id=p_tenant_id AND owner.deleted_at IS NULL;
    END IF;
    IF pg_catalog.to_regclass('module_forms') IS NOT NULL THEN
        RETURN QUERY SELECT 'source','form',row.id::text,to_jsonb(row)::text
        FROM module_forms AS row JOIN modules AS owner ON owner.id=row.module_id
        WHERE owner.tenant_id=p_tenant_id AND owner.deleted_at IS NULL;
    END IF;
    IF pg_catalog.to_regclass('module_schema_versions') IS NOT NULL THEN
        RETURN QUERY SELECT 'source','schema_version',row.id::text,to_jsonb(row)::text
        FROM module_schema_versions AS row WHERE row.tenant_id=p_tenant_id;
    END IF;
    IF pg_catalog.to_regclass('module_migration_runs') IS NOT NULL THEN
        RETURN QUERY SELECT 'source','migration_run',row.id::text,to_jsonb(row)::text
        FROM module_migration_runs AS row WHERE row.tenant_id=p_tenant_id;
    END IF;
    IF pg_catalog.to_regclass('menu_configs') IS NOT NULL THEN
        RETURN QUERY SELECT 'source','menu',row.id::text,to_jsonb(row)::text
        FROM menu_configs AS row WHERE row.tenant_id=p_tenant_id
          AND row.deleted_at IS NULL;
    END IF;
    RETURN QUERY
    SELECT 'target','installation',row.installation_id::text,to_jsonb(row)::text
      FROM catalog_installations AS row
    UNION ALL SELECT 'target','binding',row.component_id::text,to_jsonb(row)::text
      FROM catalog_component_bindings AS row
    UNION ALL SELECT 'target','dependency',
      row.installation_id::text||':'||row.dependency_package_id::text,to_jsonb(row)::text
      FROM catalog_dependency_resolutions AS row
    UNION ALL SELECT 'target','menu_definition',row.definition_id::text,to_jsonb(row)::text
      FROM catalog_menu_definitions AS row
    UNION ALL SELECT 'target','menu_version',row.version_id::text,to_jsonb(row)::text
      FROM catalog_menu_versions AS row
    UNION ALL SELECT 'target','menu_item',
      row.version_id::text||':'||row.item_id::text,to_jsonb(row)::text
      FROM catalog_menu_items AS row;
END;
$function$;
CREATE FUNCTION catalog_adoption_inventory_root_v1(
    p_tenant_id uuid, p_side text
) RETURNS TABLE(row_count bigint, root_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_side NOT IN ('source','target') THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Catalog adoption inventory side invalid';
    END IF;
    RETURN QUERY SELECT pg_catalog.count(*),
      pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
        pg_catalog.int4send(pg_catalog.octet_length(row.kind))||
        pg_catalog.convert_to(row.kind,'UTF8')||
        pg_catalog.int4send(pg_catalog.octet_length(row.identity))||
        pg_catalog.convert_to(row.identity,'UTF8')||
        pg_catalog.int4send(pg_catalog.octet_length(row.document))||
        pg_catalog.convert_to(row.document,'UTF8'),
        ''::bytea ORDER BY row.kind,row.identity),''::bytea)),'hex')
    FROM catalog_adoption_inventory_v1(p_tenant_id) AS row
    WHERE row.side=p_side;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_adoption_inventory_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION catalog_adoption_inventory_root_v1(uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION catalog_adoption_inventory_v1(uuid) TO harpy_runtime;
