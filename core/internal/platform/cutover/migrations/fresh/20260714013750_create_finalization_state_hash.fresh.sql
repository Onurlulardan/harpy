-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.finalization_state_sha256_v1(
    p_operation_id uuid, p_run_id uuid
)
RETURNS text LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
SET TimeZone TO 'UTC'
SET DateStyle TO 'ISO, YMD'
SET bytea_output TO 'hex'
SET IntervalStyle TO 'iso_8601'
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_release_run_id uuid;
BEGIN
    IF p_operation_id IS NULL OR p_run_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='finalization state identity is invalid';
    END IF;
    LOCK TABLE platform.cutover_operations,
        platform.cutover_operation_relations,
        platform.cutover_operation_checkpoint_keys,
        platform.cutover_operation_columns,
        platform.cutover_operation_procedures,
        platform.cutover_initial_copy_receipts,
        platform.cutover_batch_receipts,
        platform.cutover_copy_checkpoints,
        platform.applied_change_receipts,
        platform.cutover_change_journal,
        platform.cutover_credential_receipts,
        platform.tenant_migration_requirement_catalogs,
        platform.tenant_migration_requirements,
        platform.tenant_migration_requirement_targets,
        platform.tenant_migration_checkpoints,
        platform.tenants,
        platform.release_generations,
        platform.schema_contracts IN SHARE MODE;
    SELECT operation.* INTO STRICT v_operation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id;
    SELECT release_row.migration_run_id INTO STRICT v_release_run_id
    FROM platform.release_generations AS release_row
    WHERE release_row.release_id=v_operation.release_id
      AND release_row.environment_id=v_operation.environment_id;
    RETURN (SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        pg_catalog.jsonb_build_object(
            'operation',pg_catalog.to_jsonb(v_operation),
            'relations',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(relation)
                ORDER BY relation.ordinal) FROM platform.cutover_operation_relations AS relation
                WHERE relation.operation_id=p_operation_id AND relation.run_id=p_run_id),'[]'::jsonb),
            'checkpoint_keys',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(checkpoint_key)
                ORDER BY checkpoint_key.relation_ordinal,checkpoint_key.ordinal)
                FROM platform.cutover_operation_checkpoint_keys AS checkpoint_key
                WHERE checkpoint_key.operation_id=p_operation_id AND checkpoint_key.run_id=p_run_id),'[]'::jsonb),
            'columns',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(column_row)
                ORDER BY column_row.relation_ordinal,column_row.ordinal)
                FROM platform.cutover_operation_columns AS column_row
                WHERE column_row.operation_id=p_operation_id AND column_row.run_id=p_run_id),'[]'::jsonb),
            'procedures',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(procedure_row)
                ORDER BY procedure_row.ordinal) FROM platform.cutover_operation_procedures AS procedure_row
                WHERE procedure_row.operation_id=p_operation_id AND procedure_row.run_id=p_run_id),'[]'::jsonb),
            'initial_copy_receipt',(SELECT pg_catalog.to_jsonb(receipt)
                FROM platform.cutover_initial_copy_receipts AS receipt
                WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id),
            'copy_checkpoints',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(checkpoint)
                ORDER BY checkpoint.relation_ordinal,checkpoint.checkpoint_key)
                FROM platform.cutover_copy_checkpoints AS checkpoint
                WHERE checkpoint.operation_id=p_operation_id AND checkpoint.run_id=p_run_id),'[]'::jsonb),
            'batch_receipts',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(batch)
                ORDER BY batch.relation_ordinal,batch.checkpoint_key,batch.batch_ordinal)
                FROM platform.cutover_batch_receipts AS batch
                WHERE batch.operation_id=p_operation_id AND batch.run_id=p_run_id),'[]'::jsonb),
            'applied_changes',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(receipt)
                ORDER BY receipt.source_xid,receipt.relation_ordinal,receipt.change_ordinal)
                FROM platform.applied_change_receipts AS receipt
                WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id),'[]'::jsonb),
            'change_journal',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(journal)
                ORDER BY journal.source_xid,journal.domain)
                FROM platform.cutover_change_journal AS journal
                WHERE journal.operation_id=p_operation_id AND journal.run_id=p_run_id),'[]'::jsonb),
            'credential_receipts',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(receipt)
                ORDER BY receipt.receipt_kind,receipt.receipt_id)
                FROM platform.cutover_credential_receipts AS receipt
                WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id),'[]'::jsonb),
            'release',(SELECT pg_catalog.to_jsonb(release_row) FROM platform.release_generations AS release_row
                WHERE release_row.migration_run_id=v_release_run_id),
            'schema_contract',(SELECT pg_catalog.to_jsonb(contract_row) FROM platform.schema_contracts AS contract_row
                WHERE contract_row.component='core'),
            'tenant_catalog',(SELECT pg_catalog.to_jsonb(catalog) FROM platform.tenant_migration_requirement_catalogs AS catalog
                WHERE catalog.migration_run_id=v_release_run_id),
            'tenant_requirements',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(requirement)
                ORDER BY requirement.ordinal) FROM platform.tenant_migration_requirements AS requirement
                WHERE requirement.migration_run_id=v_release_run_id),'[]'::jsonb),
            'tenant_targets',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(target)
                ORDER BY target.ordinal) FROM platform.tenant_migration_requirement_targets AS target
                WHERE target.migration_run_id=v_release_run_id),'[]'::jsonb),
            'active_tenants',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
                'tenant_id',tenant.id::text,'schema_name',tenant.schema_name) ORDER BY tenant.id)
                FROM platform.tenants AS tenant WHERE tenant.deleted_at IS NULL),'[]'::jsonb),
            'tenant_checkpoints',COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(checkpoint)
                ORDER BY checkpoint.tenant_id,checkpoint.owner,checkpoint.logical_version)
                FROM platform.tenant_migration_checkpoints AS checkpoint
                WHERE EXISTS (SELECT 1 FROM platform.tenant_migration_requirement_targets AS target
                    WHERE target.migration_run_id=v_release_run_id AND target.tenant_id=checkpoint.tenant_id)
                  AND EXISTS (SELECT 1 FROM platform.tenant_migration_requirements AS requirement
                    WHERE requirement.migration_run_id=v_release_run_id
                      AND requirement.owner=checkpoint.owner
                      AND requirement.logical_version=checkpoint.logical_version)),'[]'::jsonb)
        )::text,'UTF8')),'hex'));
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.finalization_state_sha256_v1(uuid,uuid) FROM PUBLIC;
