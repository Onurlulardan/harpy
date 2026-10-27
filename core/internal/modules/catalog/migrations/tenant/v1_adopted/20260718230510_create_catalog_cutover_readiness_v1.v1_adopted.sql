-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_v1_adoption_projection_ready_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_id uuid
) RETURNS TABLE(final_lsn pg_lsn,evidence jsonb)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_generation bigint;
    v_expected bigint;
    v_complete bigint;
    v_final_count bigint;
    v_evidence_count bigint;
    v_existing bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog V1 adoption readiness principal denied';
    END IF;
    SELECT operation.cutover_generation INTO v_generation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
      AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
      AND operation.group_id='catalog' AND operation.domain='catalog'
      AND operation.state='quiescing'
    FOR SHARE;
    IF NOT FOUND OR NOT EXISTS (
      SELECT 1 FROM tenancy_schema_identity AS tenant
      WHERE tenant.singleton AND tenant.tenant_id=p_scope_id
        AND tenant.schema_name=pg_catalog.current_schema()
        AND tenant.origin='v1_adopted'
    ) OR NOT EXISTS (
      SELECT 1 FROM platform.cutover_groups AS fence
      WHERE fence.scope_kind='tenant' AND fence.scope_id=p_scope_id
        AND fence.group_id='catalog' AND fence.generation=v_generation
        AND fence.state='quiescing' AND fence.writer='none'
    ) OR NOT EXISTS (
      SELECT 1 FROM platform.domain_cutovers AS fence
      WHERE fence.scope_kind='tenant' AND fence.scope_id=p_scope_id
        AND fence.domain='catalog' AND fence.generation=v_generation
        AND fence.state='quiescing' AND fence.writer='none'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption projection authority not ready';
    END IF;
    SELECT pg_catalog.count(*) INTO v_expected
    FROM platform.cutover_operation_checkpoint_keys AS expected
    WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id
      AND expected.relation_ordinal=1 AND expected.checkpoint_key IN (
        'modules','module_dependencies','module_forms',
        'module_schema_versions','module_migration_runs','menu_configs'
      );
    SELECT pg_catalog.count(*),
      pg_catalog.count(DISTINCT checkpoint.final_lsn::text),
      pg_catalog.count(DISTINCT checkpoint.last_key)
    INTO v_complete,v_final_count,v_evidence_count
    FROM platform.cutover_copy_checkpoints AS checkpoint
    WHERE checkpoint.operation_id=p_operation_id AND checkpoint.run_id=p_run_id
      AND checkpoint.relation_ordinal=1 AND checkpoint.state='complete'
      AND checkpoint.final_lsn IS NOT NULL
      AND checkpoint.applied_lsn=checkpoint.final_lsn;
    IF v_expected<>6 OR (SELECT pg_catalog.count(*)
      FROM platform.cutover_operation_checkpoint_keys AS expected
      WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id)<>6 OR
      v_complete<>6 OR v_final_count<>1 OR v_evidence_count<>1 OR EXISTS (
      SELECT 1 FROM platform.cutover_operation_checkpoint_keys AS expected
      LEFT JOIN platform.cutover_copy_checkpoints AS checkpoint
        ON checkpoint.operation_id=expected.operation_id
       AND checkpoint.run_id=expected.run_id
       AND checkpoint.relation_ordinal=expected.relation_ordinal
       AND checkpoint.checkpoint_key=expected.checkpoint_key
       AND checkpoint.state='complete'
      WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id
        AND checkpoint.operation_id IS NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption checkpoints incomplete';
    END IF;
    SELECT
      (SELECT pg_catalog.count(*) FROM catalog_installations)+
      (SELECT pg_catalog.count(*) FROM catalog_component_bindings)+
      (SELECT pg_catalog.count(*) FROM catalog_dependency_resolutions)+
      (SELECT pg_catalog.count(*) FROM catalog_menu_definitions)+
      (SELECT pg_catalog.count(*) FROM catalog_menu_versions)+
      (SELECT pg_catalog.count(*) FROM catalog_menu_items)
    INTO v_existing;
    IF v_existing<>0 OR EXISTS (
      SELECT 1 FROM catalog_operation_runs AS operation
      WHERE operation.state IN (
        'PLANNED','APPLYING','WAITING','VERIFYING','ACTIVATING','CANCELLING'
      )
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption inactive target not ready';
    END IF;
    RETURN QUERY
    SELECT checkpoint.final_lsn,checkpoint.last_key
    FROM platform.cutover_copy_checkpoints AS checkpoint
    WHERE checkpoint.operation_id=p_operation_id AND checkpoint.run_id=p_run_id
    ORDER BY checkpoint.checkpoint_key LIMIT 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_v1_adoption_projection_ready_v1(
    uuid,uuid,uuid
) FROM PUBLIC;

