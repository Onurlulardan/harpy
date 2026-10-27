-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_adoption_blockers_v1(p_tenant_id uuid)
RETURNS TABLE(blocker text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER NOT IN ('forgecrud_v2_runtime','forgecrud_adopter') OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR NOT EXISTS (
        SELECT 1 FROM tenancy_schema_identity AS tenant
        WHERE tenant.singleton AND tenant.tenant_id=p_tenant_id
          AND tenant.schema_name=pg_catalog.current_schema()
    ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog adoption blocker authority denied';
    END IF;
    IF pg_catalog.to_regclass('modules') IS NULL THEN
        RETURN;
    END IF;
    IF pg_catalog.to_regclass('module_dependencies') IS NULL OR
       pg_catalog.to_regclass('module_forms') IS NULL OR
       pg_catalog.to_regclass('module_schema_versions') IS NULL OR
       pg_catalog.to_regclass('module_migration_runs') IS NULL OR
       pg_catalog.to_regclass('menu_configs') IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog adoption legacy inventory is incomplete';
    END IF;
    RETURN QUERY
    SELECT value FROM (
      SELECT 'ORPHAN_DEPENDENCY:'||dependency.id::text AS value
      FROM module_dependencies AS dependency
      LEFT JOIN modules AS owner ON owner.id=dependency.module_id
      LEFT JOIN modules AS wanted ON wanted.id=dependency.depends_on_module_id
      WHERE owner.id IS NULL OR wanted.id IS NULL OR
        owner.tenant_id<>p_tenant_id OR wanted.tenant_id<>p_tenant_id OR
        owner.deleted_at IS NOT NULL OR wanted.deleted_at IS NOT NULL
      UNION ALL
      SELECT 'ORPHAN_FORM:'||form_row.id::text
      FROM module_forms AS form_row
      LEFT JOIN modules AS owner ON owner.id=form_row.module_id
      WHERE owner.id IS NULL OR owner.tenant_id<>p_tenant_id OR
        owner.deleted_at IS NOT NULL OR form_row.form_config_id IS NULL
      UNION ALL
      SELECT 'UNRESOLVED_MENU:'||menu.id::text FROM menu_configs AS menu
      WHERE menu.tenant_id=p_tenant_id AND menu.deleted_at IS NULL
        AND pg_catalog.jsonb_typeof(menu.config) NOT IN ('object','array')
      UNION ALL
      SELECT CASE registry.blocker
        WHEN 'UNKNOWN_PACKAGE' THEN 'UNKNOWN_MODULE'
        ELSE registry.blocker END||':'||module.id::text
      FROM modules AS module
      CROSS JOIN platform.catalog_adoption_registry_blocker_v1(
        module.id,module.slug,module.version,NULL
      ) AS registry(blocker)
      WHERE module.tenant_id=p_tenant_id AND module.deleted_at IS NULL
        AND registry.blocker IS NOT NULL
      UNION ALL
      SELECT CASE registry.blocker
        WHEN 'UNKNOWN_PACKAGE' THEN 'UNKNOWN_MANIFEST_PACKAGE'
        WHEN 'UNKNOWN_VERSION' THEN 'UNKNOWN_MANIFEST_VERSION'
        ELSE registry.blocker END||':'||source.id::text
      FROM module_schema_versions AS source
      CROSS JOIN platform.catalog_adoption_registry_blocker_v1(
        NULL,source.module_slug,source.module_version,source.manifest_hash
      ) AS registry(blocker)
      WHERE source.tenant_id=p_tenant_id AND registry.blocker IS NOT NULL
    ) AS blockers ORDER BY value;
END;
$function$;
CREATE FUNCTION catalog_adoption_blocker_sha256_v1(p_tenant_id uuid)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_digest text;
BEGIN
    SELECT pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
      pg_catalog.int4send(pg_catalog.octet_length(row.blocker))||
      pg_catalog.convert_to(row.blocker,'UTF8'),
      ''::bytea ORDER BY row.blocker),''::bytea)),'hex')
    INTO v_digest FROM catalog_adoption_blockers_v1(p_tenant_id) AS row;
    RETURN v_digest;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_adoption_blockers_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION catalog_adoption_blocker_sha256_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION catalog_adoption_blockers_v1(uuid) TO forgecrud_runtime;
