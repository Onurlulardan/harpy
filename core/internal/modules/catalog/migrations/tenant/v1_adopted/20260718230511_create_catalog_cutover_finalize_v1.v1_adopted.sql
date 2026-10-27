-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_v1_adoption_finalize_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_id uuid
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_kind text;
    v_stage catalog_v1_adoption_stage%ROWTYPE;
    v_generation bigint;
    v_plan text;
    v_transform text;
    v_lease text;
    v_final_lsn pg_lsn;
    v_evidence jsonb;
    v_source_count bigint;
    v_source_root text;
    v_target_count bigint;
    v_target_root text;
    v_live_count bigint;
    v_live_root text;
    v_blocker text;
    v_consumer text;
    v_acceptance text;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog V1 adoption finalization principal denied';
    END IF;
    v_now:=pg_catalog.date_trunc(
      'microseconds',pg_catalog.transaction_timestamp()
    );
    SELECT ready.final_lsn,ready.evidence INTO STRICT v_final_lsn,v_evidence
    FROM catalog_v1_adoption_projection_ready_v1(
      p_operation_id,p_run_id,p_scope_id
    ) AS ready;
    SELECT operation.cutover_generation,operation.plan_sha256,
      operation.transform_manifest_sha256,operation.lease_manifest_sha256
    INTO STRICT v_generation,v_plan,v_transform,v_lease
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
    FOR SHARE;
    SELECT pg_catalog.count(*),
      pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
        pg_catalog.int4send(pg_catalog.octet_length(source.kind))||
        pg_catalog.convert_to(source.kind,'UTF8')||
        pg_catalog.int4send(pg_catalog.octet_length(source.identity))||
        pg_catalog.convert_to(source.identity,'UTF8')||
        pg_catalog.int4send(pg_catalog.octet_length(source.document))||
        pg_catalog.convert_to(source.document,'UTF8'),
        ''::bytea ORDER BY source.kind,source.identity),''::bytea)),'hex')
    INTO v_source_count,v_source_root FROM (
      SELECT CASE stage.source_kind
        WHEN 'modules' THEN 'module'
        WHEN 'module_dependencies' THEN 'dependency'
        WHEN 'module_forms' THEN 'form'
        WHEN 'module_schema_versions' THEN 'schema_version'
        WHEN 'module_migration_runs' THEN 'migration_run'
        WHEN 'menu_configs' THEN 'menu' END AS kind,
        stage.source_identity AS identity,
        pg_catalog.min(stage.source_document::text) AS document
      FROM catalog_v1_adoption_stage AS stage
      WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
        AND NOT stage.removed
      GROUP BY stage.source_kind,stage.source_identity
    ) AS source;
    SELECT pg_catalog.count(*),
      pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
        pg_catalog.int4send(pg_catalog.octet_length(stage.target_kind))||
        pg_catalog.convert_to(stage.target_kind,'UTF8')||
        pg_catalog.int4send(pg_catalog.octet_length(stage.target_identity))||
        pg_catalog.convert_to(stage.target_identity,'UTF8')||
        pg_catalog.int4send(pg_catalog.octet_length(stage.target_record::text))||
        pg_catalog.convert_to(stage.target_record::text,'UTF8'),
        ''::bytea ORDER BY stage.target_kind,stage.target_identity),''::bytea)),'hex')
    INTO v_target_count,v_target_root
    FROM catalog_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed AND stage.target_kind<>'discard';
    SELECT root.row_count,root.root_sha256 INTO STRICT v_live_count,v_live_root
    FROM catalog_adoption_inventory_root_v1(p_scope_id,'source') AS root;
    v_blocker:=catalog_adoption_blocker_sha256_v1(p_scope_id);
    v_consumer:=catalog_v1_adoption_consumer_sha256_v1(p_scope_id);
    IF EXISTS (SELECT 1 FROM catalog_adoption_blockers_v1(p_scope_id)) OR
       v_source_count IS DISTINCT FROM v_live_count OR
       v_source_root IS DISTINCT FROM v_live_root OR
       (v_evidence->>'source_count')::bigint IS DISTINCT FROM v_source_count OR
       v_evidence->>'source_root_sha256' IS DISTINCT FROM v_source_root OR
       (v_evidence->>'target_count')::bigint IS DISTINCT FROM v_target_count OR
       v_evidence->>'target_root_sha256' IS DISTINCT FROM v_target_root OR
       v_evidence->>'blocker_sha256' IS DISTINCT FROM v_blocker OR
       v_evidence->>'consumer_sha256' IS DISTINCT FROM v_consumer THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption inventory evidence differs';
    END IF;
    FOREACH v_kind IN ARRAY ARRAY[
      'installation','binding','dependency',
      'menu_definition','menu_version','menu_item'
    ]::text[] LOOP
      FOR v_stage IN
        SELECT stage.* FROM catalog_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
          AND NOT stage.removed AND stage.target_kind=v_kind
        ORDER BY stage.target_identity
      LOOP
        IF v_kind IN ('installation','binding','dependency') THEN
          PERFORM catalog_v1_adoption_project_core_v1(
            v_stage.target_kind,v_stage.target_record
          );
        ELSE
          PERFORM catalog_v1_adoption_project_menu_v1(
            v_stage.target_kind,v_stage.target_record
          );
        END IF;
      END LOOP;
    END LOOP;
    FOR v_stage IN
      SELECT stage.* FROM catalog_v1_adoption_stage AS stage
      WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
        AND NOT stage.removed AND stage.target_kind='menu_definition'
      ORDER BY stage.target_identity
    LOOP
      PERFORM catalog_v1_adoption_activate_menu_v1(v_stage.target_record);
    END LOOP;
    SELECT root.row_count,root.root_sha256 INTO STRICT v_live_count,v_live_root
    FROM catalog_adoption_inventory_root_v1(p_scope_id,'target') AS root;
    IF v_live_count IS DISTINCT FROM v_target_count OR
       v_live_root IS DISTINCT FROM v_target_root THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption target projection differs';
    END IF;
    INSERT INTO catalog_adoption_runs(
      run_id,source_manifest_sha256,transform_sha256,source_count,target_count,
      source_root_sha256,target_root_sha256,state,blocker_code,
      cutover_generation,created_at,verified_at
    ) VALUES (
      p_run_id,v_plan,v_transform,v_source_count,v_target_count,
      v_source_root,v_target_root,'VERIFIED',NULL,v_generation,v_now,v_now
    ) ON CONFLICT DO NOTHING;
    PERFORM 1 FROM catalog_adoption_runs AS run
    WHERE run.run_id=p_run_id AND run.source_manifest_sha256=v_plan
      AND run.transform_sha256=v_transform AND run.source_count=v_source_count
      AND run.target_count=v_target_count AND run.source_root_sha256=v_source_root
      AND run.target_root_sha256=v_target_root AND run.state='VERIFIED'
      AND run.blocker_code IS NULL AND run.cutover_generation=v_generation
      AND run.created_at=v_now AND run.verified_at=v_now;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption run replay differs';
    END IF;
    v_acceptance:=pg_catalog.encode(pg_catalog.sha256(
      pg_catalog.convert_to('catalog-v1-adoption-acceptance.v1','UTF8')||
      pg_catalog.decode('00','hex')||pg_catalog.uuid_send(p_operation_id)||
      pg_catalog.decode('00','hex')||pg_catalog.uuid_send(p_run_id)||
      pg_catalog.decode('00','hex')||pg_catalog.uuid_send(p_scope_id)||
      pg_catalog.decode('00','hex')||pg_catalog.int8send(v_generation)||
      pg_catalog.decode('00','hex')||pg_catalog.convert_to(v_plan,'UTF8')||
      pg_catalog.decode('00','hex')||pg_catalog.convert_to(v_transform,'UTF8')||
      pg_catalog.decode('00','hex')||pg_catalog.convert_to(v_lease,'UTF8')||
      pg_catalog.decode('00','hex')||pg_catalog.convert_to(v_final_lsn::text,'UTF8')||
      pg_catalog.decode('00','hex')||
      pg_catalog.convert_to(v_evidence->>'evidence_sha256','UTF8')
    ),'hex');
    INSERT INTO catalog_v1_adoption_receipts(
      operation_id,run_id,tenant_id,cutover_generation,plan_sha256,
      transform_sha256,lease_manifest_sha256,final_lsn,source_count,
      source_root_sha256,target_count,target_root_sha256,blocker_sha256,
      consumer_sha256,evidence_sha256,acceptance_sha256,recorded_at
    ) VALUES (
      p_operation_id,p_run_id,p_scope_id,v_generation,v_plan,v_transform,v_lease,
      v_final_lsn,v_source_count,v_source_root,v_target_count,v_target_root,
      v_blocker,v_consumer,v_evidence->>'evidence_sha256',v_acceptance,v_now
    );
    RETURN v_target_root;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_v1_adoption_finalize_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
