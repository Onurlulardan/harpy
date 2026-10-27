-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_verify_new_tenant_v1(
    p_release_id uuid,
    p_environment_id uuid,
    p_database_origin text,
    p_tenant_id uuid,
    p_migration_manifest_sha256 text
)
RETURNS TABLE (
    verified boolean,
    catalog_sha256 text,
    requirement_count integer
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_migration_run_id uuid;
    v_catalog_sha256 text;
    v_requirement_count integer;
BEGIN
    LOCK TABLE platform.release_generations,
        platform.tenant_migration_requirement_catalogs,
        platform.tenant_migration_requirements,
        platform.tenant_migration_checkpoints IN SHARE MODE;

    SELECT catalog.migration_run_id, catalog.catalog_sha256,
           catalog.requirement_count
    INTO v_migration_run_id, v_catalog_sha256, v_requirement_count
    FROM platform.tenant_migration_requirement_catalogs AS catalog
    JOIN platform.release_generations AS release_row
      ON release_row.migration_run_id = catalog.migration_run_id
     AND release_row.release_id = catalog.release_id
     AND release_row.environment_id = catalog.environment_id
     AND release_row.database_origin = catalog.database_origin
     AND release_row.migration_manifest_sha256 =
         catalog.migration_manifest_sha256
    WHERE catalog.release_id = p_release_id
      AND catalog.environment_id = p_environment_id
      AND catalog.database_origin = p_database_origin
      AND catalog.migration_manifest_sha256 = p_migration_manifest_sha256
      AND catalog.contract_version =
          'tenant-migration-requirement-catalog.v2';

    IF NOT FOUND OR v_requirement_count < 1 THEN
        RETURN QUERY SELECT false, NULL::text, 0;
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM platform.tenant_migration_requirements AS requirement
        WHERE requirement.migration_run_id = v_migration_run_id
          AND NOT EXISTS (
              SELECT 1
              FROM platform.tenant_migration_checkpoints AS checkpoint
              WHERE checkpoint.tenant_id = p_tenant_id
                AND checkpoint.owner = requirement.owner
                AND checkpoint.logical_version = requirement.logical_version
                AND checkpoint.database_origin = requirement.database_origin
                AND checkpoint.selected_path = requirement.selected_path
                AND checkpoint.selected_sha256 = requirement.selected_sha256
                AND checkpoint.state = 'succeeded'
          )
    ) OR (
        SELECT pg_catalog.count(*)
        FROM platform.tenant_migration_requirements AS requirement
        WHERE requirement.migration_run_id = v_migration_run_id
    ) IS DISTINCT FROM v_requirement_count::bigint THEN
        RETURN QUERY SELECT false, v_catalog_sha256, v_requirement_count;
        RETURN;
    END IF;

    RETURN QUERY SELECT true, v_catalog_sha256, v_requirement_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_migrator_verify_new_tenant_v1(
    uuid, uuid, text, uuid, text
) FROM PUBLIC;
