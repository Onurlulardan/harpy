-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_verify_tenant_requirement_catalog_v5(
    p_migration_run_id uuid,
    p_release_id uuid,
    p_environment_id uuid,
    p_database_origin text
)
RETURNS TABLE (
    verified boolean, catalog_sha256 text,
    requirement_count integer, target_count integer
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_catalog_sha256 text;
    v_requirement_count integer;
    v_target_count integer;
    v_migration_sha256 text;
    v_document jsonb;
    v_requirements jsonb;
    v_targets jsonb;
    v_live_targets jsonb;
BEGIN
    LOCK TABLE platform.release_generations,
        platform.tenant_migration_requirement_catalogs,
        platform.tenant_migration_requirements,
        platform.tenant_migration_requirement_targets,
        platform.tenant_migration_supersessions,
        platform.tenants,
        platform.tenant_migration_checkpoints IN SHARE MODE;

    SELECT catalog.catalog_sha256,catalog.requirement_count,
           catalog.target_count,catalog.migration_manifest_sha256,
           pg_catalog.convert_from(catalog.canonical_document,'UTF8')::jsonb
    INTO v_catalog_sha256,v_requirement_count,v_target_count,
         v_migration_sha256,v_document
    FROM platform.tenant_migration_requirement_catalogs AS catalog
    JOIN platform.release_generations AS release_row
      ON release_row.migration_run_id = catalog.migration_run_id
     AND release_row.release_id = catalog.release_id
     AND release_row.environment_id = catalog.environment_id
     AND release_row.database_origin = catalog.database_origin
     AND release_row.migration_manifest_sha256 =
         catalog.migration_manifest_sha256
    WHERE catalog.migration_run_id = p_migration_run_id
      AND catalog.release_id = p_release_id
      AND catalog.environment_id = p_environment_id
      AND catalog.database_origin = p_database_origin
      AND catalog.contract_version =
          'tenant-migration-requirement-catalog.v2';
    IF NOT FOUND THEN
        RETURN QUERY SELECT false,NULL::text,0,0;
        RETURN;
    END IF;

    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'ordinal',requirement.ordinal,'owner',requirement.owner,
        'logical_version',requirement.logical_version,
        'database_origin',requirement.database_origin,
        'selected_path',requirement.selected_path,
        'selected_sha256',requirement.selected_sha256
    ) ORDER BY requirement.ordinal),'[]'::jsonb)
    INTO v_requirements
    FROM platform.tenant_migration_requirements AS requirement
    WHERE requirement.migration_run_id = p_migration_run_id;

    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'ordinal',target.ordinal,'tenant_id',target.tenant_id::text,
        'schema_name',target.schema_name
    ) ORDER BY target.ordinal),'[]'::jsonb)
    INTO v_targets
    FROM platform.tenant_migration_requirement_targets AS target
    WHERE target.migration_run_id = p_migration_run_id;

    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'ordinal',live_target.ordinal,'tenant_id',live_target.tenant_id,
        'schema_name',live_target.schema_name
    ) ORDER BY live_target.ordinal),'[]'::jsonb)
    INTO v_live_targets
    FROM (
        SELECT pg_catalog.row_number() OVER (ORDER BY tenant.id) AS ordinal,
               tenant.id::text AS tenant_id,tenant.schema_name
        FROM platform.tenants AS tenant WHERE tenant.deleted_at IS NULL
    ) AS live_target;

    IF pg_catalog.jsonb_typeof(v_document) IS DISTINCT FROM 'object'
       OR (SELECT pg_catalog.count(*) FROM
           pg_catalog.jsonb_object_keys(v_document)) IS DISTINCT FROM 8::bigint
       OR v_document->>'contract_version' IS DISTINCT FROM
          'tenant-migration-requirement-catalog.v2'
       OR v_document->>'migration_run_id' IS DISTINCT FROM
          p_migration_run_id::text
       OR v_document->>'release_id' IS DISTINCT FROM p_release_id::text
       OR v_document->>'environment_id' IS DISTINCT FROM p_environment_id::text
       OR v_document->>'database_origin' IS DISTINCT FROM p_database_origin
       OR pg_catalog.jsonb_typeof(v_document->'contract_version') <> 'string'
       OR pg_catalog.jsonb_typeof(v_document->'migration_run_id') <> 'string'
       OR pg_catalog.jsonb_typeof(v_document->'release_id') <> 'string'
       OR pg_catalog.jsonb_typeof(v_document->'environment_id') <> 'string'
       OR pg_catalog.jsonb_typeof(v_document->'database_origin') <> 'string'
       OR pg_catalog.jsonb_typeof(
          v_document->'migration_manifest_sha256') <> 'string'
       OR v_document->>'migration_manifest_sha256' IS DISTINCT FROM
          v_migration_sha256
       OR pg_catalog.jsonb_typeof(v_document->'requirements') <> 'array'
       OR pg_catalog.jsonb_typeof(v_document->'targets') <> 'array'
       OR v_document->'requirements' IS DISTINCT FROM v_requirements
       OR v_document->'targets' IS DISTINCT FROM v_targets
       OR v_targets IS DISTINCT FROM v_live_targets
       OR pg_catalog.jsonb_array_length(v_requirements) IS DISTINCT FROM
          v_requirement_count
       OR pg_catalog.jsonb_array_length(v_targets) IS DISTINCT FROM
          v_target_count THEN
        RETURN QUERY SELECT false,v_catalog_sha256,
            v_requirement_count,v_target_count;
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM platform.tenant_migration_requirement_targets AS target
        CROSS JOIN platform.tenant_migration_requirements AS requirement
        WHERE target.migration_run_id = p_migration_run_id
          AND requirement.migration_run_id = p_migration_run_id
          AND NOT platform.platform_migrator_tenant_checkpoint_satisfies_v3(
              target.tenant_id,requirement.owner,
              requirement.logical_version,requirement.database_origin,
              requirement.selected_path,requirement.selected_sha256
          )
    ) THEN
        RETURN QUERY SELECT false,v_catalog_sha256,
            v_requirement_count,v_target_count;
        RETURN;
    END IF;

    RETURN QUERY SELECT true,v_catalog_sha256,
        v_requirement_count,v_target_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_migrator_verify_tenant_requirement_catalog_v5(
        uuid,uuid,uuid,text
    )
FROM PUBLIC;
