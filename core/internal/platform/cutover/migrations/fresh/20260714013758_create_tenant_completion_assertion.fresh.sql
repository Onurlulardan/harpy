-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.assert_tenant_completion_v1(
    p_operation_id uuid, p_run_id uuid
)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_catalog platform.tenant_migration_requirement_catalogs%ROWTYPE;
    v_catalog_document jsonb;
    v_requirements jsonb;
    v_targets jsonb;
    v_live_targets jsonb;
    v_release_run_id uuid;
    v_database_origin text;
BEGIN
    IF SESSION_USER NOT IN ('harpy_cluster_admin','harpy_cutover_finalizer') OR
       p_operation_id IS NULL OR p_run_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover tenant completion input is invalid';
    END IF;
    SELECT operation.* INTO STRICT v_operation FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id AND operation.state='quiescing';
    SELECT release_row.migration_run_id,release_row.database_origin INTO STRICT v_release_run_id,v_database_origin
    FROM platform.release_generations AS release_row JOIN platform.schema_contracts AS contract_row
      ON contract_row.component='core' AND contract_row.verified_run_id=release_row.migration_run_id
     AND contract_row.database_generation=release_row.database_generation
     AND contract_row.minimum_runtime_generation=release_row.minimum_runtime_generation
     AND contract_row.migration_manifest_sha256=release_row.migration_manifest_sha256
     AND contract_row.extension_manifest_sha256=release_row.extension_manifest_sha256 AND contract_row.status='ready'
    WHERE release_row.release_id=v_operation.release_id AND release_row.environment_id=v_operation.environment_id;
    SELECT catalog.* INTO STRICT v_catalog FROM platform.tenant_migration_requirement_catalogs AS catalog
    WHERE catalog.migration_run_id=v_release_run_id AND catalog.release_id=v_operation.release_id
      AND catalog.environment_id=v_operation.environment_id AND catalog.database_origin=v_database_origin;
    v_catalog_document:=pg_catalog.convert_from(v_catalog.canonical_document,'UTF8')::jsonb;
    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object('ordinal',requirement.ordinal,
      'owner',requirement.owner,'logical_version',requirement.logical_version,'database_origin',requirement.database_origin,
      'selected_path',requirement.selected_path,'selected_sha256',requirement.selected_sha256) ORDER BY requirement.ordinal),'[]'::jsonb)
    INTO v_requirements FROM platform.tenant_migration_requirements AS requirement WHERE requirement.migration_run_id=v_release_run_id;
    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object('ordinal',target.ordinal,
      'tenant_id',target.tenant_id::text,'schema_name',target.schema_name) ORDER BY target.ordinal),'[]'::jsonb)
    INTO v_targets FROM platform.tenant_migration_requirement_targets AS target WHERE target.migration_run_id=v_release_run_id;
    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object('ordinal',live.ordinal,
      'tenant_id',live.tenant_id,'schema_name',live.schema_name) ORDER BY live.ordinal),'[]'::jsonb)
    INTO v_live_targets FROM (SELECT pg_catalog.row_number() OVER (ORDER BY tenant.id) AS ordinal,
      tenant.id::text AS tenant_id,tenant.schema_name FROM platform.tenants AS tenant WHERE tenant.deleted_at IS NULL) AS live;
    IF (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_catalog_document)) IS DISTINCT FROM 8::bigint OR
       (v_catalog_document ?& ARRAY['contract_version','migration_run_id','release_id','environment_id','database_origin','migration_manifest_sha256','requirements','targets']) IS DISTINCT FROM true OR
       v_catalog_document->>'contract_version' IS DISTINCT FROM 'tenant-migration-requirement-catalog.v2' OR
       v_catalog_document->>'migration_run_id' IS DISTINCT FROM v_release_run_id::text OR
       v_catalog_document->>'release_id' IS DISTINCT FROM v_operation.release_id::text OR
       v_catalog_document->>'environment_id' IS DISTINCT FROM v_operation.environment_id::text OR
       v_catalog_document->>'database_origin' IS DISTINCT FROM v_database_origin OR
       v_catalog_document->>'migration_manifest_sha256' IS DISTINCT FROM v_catalog.migration_manifest_sha256 OR
       v_catalog_document->'requirements' IS DISTINCT FROM v_requirements OR v_catalog_document->'targets' IS DISTINCT FROM v_targets OR
       v_targets IS DISTINCT FROM v_live_targets OR pg_catalog.jsonb_array_length(v_requirements) IS DISTINCT FROM v_catalog.requirement_count OR
       pg_catalog.jsonb_array_length(v_targets) IS DISTINCT FROM v_catalog.target_count OR
       EXISTS (SELECT 1 FROM platform.tenant_migration_requirement_targets AS target
          CROSS JOIN platform.tenant_migration_requirements AS requirement
          WHERE target.migration_run_id=v_release_run_id AND requirement.migration_run_id=v_release_run_id
            AND NOT EXISTS (SELECT 1 FROM platform.tenant_migration_checkpoints AS checkpoint
              WHERE checkpoint.tenant_id=target.tenant_id AND checkpoint.owner=requirement.owner
                AND checkpoint.logical_version=requirement.logical_version AND checkpoint.database_origin=requirement.database_origin
                AND checkpoint.selected_path=requirement.selected_path AND checkpoint.selected_sha256=requirement.selected_sha256
                AND checkpoint.state='succeeded')) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover tenant completion authority differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.assert_tenant_completion_v1(uuid,uuid) FROM PUBLIC;
