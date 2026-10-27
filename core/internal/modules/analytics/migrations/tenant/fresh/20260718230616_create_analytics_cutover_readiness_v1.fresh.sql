-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_projection_ready_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_id uuid
) RETURNS TABLE(final_lsn pg_lsn,evidence jsonb)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_generation bigint;
    v_target_count bigint;
    v_public_counts jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
      RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='Analytics V1 adoption readiness principal denied';
    END IF;
    SELECT operation.cutover_generation INTO v_generation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
      AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
      AND operation.group_id='analytics' AND operation.domain='analytics'
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
        AND fence.group_id='analytics' AND fence.generation=v_generation
        AND fence.state='quiescing' AND fence.writer='none'
    ) OR NOT EXISTS (
      SELECT 1 FROM platform.domain_cutovers AS fence
      WHERE fence.scope_kind='tenant' AND fence.scope_id=p_scope_id
        AND fence.domain='analytics' AND fence.generation=v_generation
        AND fence.state='quiescing' AND fence.writer='none'
    ) THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption projection authority not ready';
    END IF;
    IF (SELECT pg_catalog.count(*)
        FROM platform.cutover_operation_checkpoint_keys AS expected
        WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id)<>
       1 OR NOT EXISTS (
      SELECT 1 FROM platform.cutover_operation_checkpoint_keys AS expected
      WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id
        AND expected.relation_ordinal=1 AND expected.ordinal=1
        AND expected.checkpoint_key='openapi_endpoints'
    ) OR NOT EXISTS (
      SELECT 1 FROM platform.cutover_copy_checkpoints AS checkpoint
      WHERE checkpoint.operation_id=p_operation_id AND checkpoint.run_id=p_run_id
        AND checkpoint.relation_ordinal=1
        AND checkpoint.checkpoint_key='openapi_endpoints'
        AND checkpoint.state='complete' AND checkpoint.final_lsn IS NOT NULL
        AND checkpoint.applied_lsn=checkpoint.final_lsn
        AND pg_catalog.jsonb_typeof(checkpoint.last_key)='object'
    ) THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption checkpoint incomplete';
    END IF;
    v_public_counts:=platform.analytics_v1_adoption_public_bridge_v1(
      'count',p_scope_id,NULL,NULL);
    SELECT
      (SELECT pg_catalog.count(*) FROM analytics_saved_queries)+
      (SELECT pg_catalog.count(*) FROM analytics_saved_query_versions)+
      (SELECT pg_catalog.count(*) FROM analytics_query_dependencies)+
      (SELECT pg_catalog.count(*) FROM analytics_publication_releases)+
      (SELECT pg_catalog.count(*) FROM analytics_publication_approvals)+
      (v_public_counts->>'publication_count')::bigint+
      (v_public_counts->>'route_count')::bigint+
      (SELECT pg_catalog.count(*) FROM analytics_v1_adoption_items
       WHERE run_id=p_run_id)+
      (SELECT pg_catalog.count(*) FROM analytics_v1_adoption_acceptance_receipts
       WHERE run_id=p_run_id)+
      (SELECT pg_catalog.count(*) FROM analytics_v1_adoption_receipts
       WHERE operation_id=p_operation_id AND run_id=p_run_id)
    INTO v_target_count;
    IF v_target_count<>0 THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption inactive target not ready';
    END IF;
    RETURN QUERY SELECT checkpoint.final_lsn,checkpoint.last_key
    FROM platform.cutover_copy_checkpoints AS checkpoint
    WHERE checkpoint.operation_id=p_operation_id AND checkpoint.run_id=p_run_id
      AND checkpoint.relation_ordinal=1
      AND checkpoint.checkpoint_key='openapi_endpoints';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_projection_ready_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
