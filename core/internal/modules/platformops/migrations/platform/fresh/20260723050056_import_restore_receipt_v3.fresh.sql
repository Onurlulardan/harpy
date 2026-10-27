-- +goose Up
ALTER TABLE platform.platformops_restore_receipt_imports
    ADD COLUMN drain_ack_sha256 text
        CHECK (drain_ack_sha256 IS NULL OR
               drain_ack_sha256 ~ '^[0-9a-f]{64}$'),
    ADD COLUMN switch_ack_sha256 text
        CHECK (switch_ack_sha256 IS NULL OR
               switch_ack_sha256 ~ '^[0-9a-f]{64}$'),
    ADD COLUMN control_database_locator_sha256 text
        CHECK (control_database_locator_sha256 IS NULL OR
               control_database_locator_sha256 ~ '^[0-9a-f]{64}$'),
    ADD COLUMN scratch_database_locator_sha256 text
        CHECK (scratch_database_locator_sha256 IS NULL OR
               scratch_database_locator_sha256 ~ '^[0-9a-f]{64}$');

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_import_restore_receipt_v3(
    p_operation_id uuid,p_recovery_set_id uuid,p_environment_id uuid,
    p_plan_sha256 text,p_manifest_sha256 text,p_receipt_sha256 text,
    p_receipt_bytes bytea,p_previous_database_system_id uuid,
    p_new_database_system_id uuid,p_object_merkle_sha256 text,
    p_locator_remap_sha256 text,p_safety_evidence_sha256 text,
    p_drain_ack_sha256 text,p_switch_ack_sha256 text,
    p_control_database_locator_sha256 text,
    p_scratch_database_locator_sha256 text,p_component_inventory jsonb
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.statement_timestamp();
    v_receipt jsonb;
    v_manifest_inventory jsonb;
    v_existing platform.platformops_restore_receipt_imports%ROWTYPE;
    v_lineage_generation bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-restore'
       OR p_operation_id IS NULL OR p_recovery_set_id IS NULL
       OR p_environment_id IS NULL OR p_previous_database_system_id IS NULL
       OR p_new_database_system_id IS NULL
       OR p_previous_database_system_id=p_new_database_system_id
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$'
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$'
       OR p_receipt_sha256 !~ '^[0-9a-f]{64}$'
       OR p_object_merkle_sha256 !~ '^[0-9a-f]{64}$'
       OR p_locator_remap_sha256 !~ '^[0-9a-f]{64}$'
       OR p_safety_evidence_sha256 !~ '^[0-9a-f]{64}$'
       OR p_drain_ack_sha256 !~ '^[0-9a-f]{64}$'
       OR p_switch_ack_sha256 !~ '^[0-9a-f]{64}$'
       OR p_control_database_locator_sha256 !~ '^[0-9a-f]{64}$'
       OR p_scratch_database_locator_sha256 !~ '^[0-9a-f]{64}$'
       OR p_control_database_locator_sha256=
          p_scratch_database_locator_sha256
       OR pg_catalog.octet_length(p_receipt_bytes) NOT BETWEEN 1 AND 2097152
       OR p_receipt_sha256<>pg_catalog.encode(
           pg_catalog.sha256(p_receipt_bytes),'hex'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore receipt import v3 denied';
    END IF;
    SELECT pg_catalog.convert_from(manifest_bytes,'UTF8')::jsonb
           #> '{payload,components}'
    INTO v_manifest_inventory
    FROM platform.platformops_recovery_sets
    WHERE recovery_set_id=p_recovery_set_id
      AND environment_id=p_environment_id
      AND manifest_sha256=p_manifest_sha256
      AND state='protected';
    IF NOT FOUND OR v_manifest_inventory IS DISTINCT FROM
       p_component_inventory THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore receipt component import v3 denied';
    END IF;
    v_receipt := pg_catalog.convert_from(p_receipt_bytes,'UTF8')::jsonb;
    IF v_receipt->>'schema' IS DISTINCT FROM
          'harpy-recovery-artifact.v1'
       OR v_receipt->>'purpose' IS DISTINCT FROM 'restore_receipt'
       OR v_receipt->>'environment_id' IS DISTINCT FROM
          p_environment_id::text
       OR v_receipt->>'operation_id' IS DISTINCT FROM p_operation_id::text
       OR v_receipt->>'artifact_id' IS DISTINCT FROM p_operation_id::text
       OR v_receipt->>'algorithm' IS DISTINCT FROM 'Ed25519'
       OR v_receipt->>'payload_sha256' !~ '^[0-9a-f]{64}$'
       OR pg_catalog.octet_length(v_receipt->>'signature')<1
       OR (v_receipt->>'issued_at')::timestamptz>v_now
       OR (v_receipt->>'expires_at')::timestamptz<=v_now
       OR v_receipt#>>'{payload,schema}' IS DISTINCT FROM
          'restore-receipt-payload.v2'
       OR v_receipt#>>'{payload,operation_id}' IS DISTINCT FROM
          p_operation_id::text
       OR v_receipt#>>'{payload,recovery_set_id}' IS DISTINCT FROM
          p_recovery_set_id::text
       OR v_receipt#>>'{payload,plan_sha256}' IS DISTINCT FROM p_plan_sha256
       OR v_receipt#>>'{payload,manifest_sha256}' IS DISTINCT FROM
          p_manifest_sha256
       OR v_receipt#>>'{payload,identity,environment_id}' IS DISTINCT FROM
          p_environment_id::text
       OR v_receipt#>>'{payload,previous_database_system_id}' IS DISTINCT FROM
          p_previous_database_system_id::text
       OR v_receipt#>>'{payload,new_database_system_id}' IS DISTINCT FROM
          p_new_database_system_id::text
       OR v_receipt#>>'{payload,previous_database_locator_sha256}'
          IS DISTINCT FROM p_control_database_locator_sha256
       OR v_receipt#>>'{payload,new_database_locator_sha256}'
          IS DISTINCT FROM p_scratch_database_locator_sha256
       OR v_receipt#>>'{payload,drain_ack_sha256}' IS DISTINCT FROM
          p_drain_ack_sha256
       OR v_receipt#>>'{payload,switch_ack_sha256}' IS DISTINCT FROM
          p_switch_ack_sha256
       OR v_receipt#>>'{payload,object_merkle_sha256}' IS DISTINCT FROM
          p_object_merkle_sha256
       OR v_receipt#>>'{payload,locator_remap_sha256}' IS DISTINCT FROM
          p_locator_remap_sha256
       OR v_receipt#>>'{payload,safety_evidence_sha256}' IS DISTINCT FROM
          p_safety_evidence_sha256
       OR (v_receipt#>>'{payload,seeded}')::boolean
       OR (v_receipt#>>'{payload,owner_changed}')::boolean
       OR (v_receipt#>>'{payload,tenant_replaced}')::boolean
       OR v_receipt#>>'{payload,switch_outcome}' IS DISTINCT FROM
          'switched_same_environment' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore receipt import v3 evidence denied';
    END IF;
    SELECT * INTO v_existing
    FROM platform.platformops_restore_receipt_imports
    WHERE operation_id=p_operation_id FOR UPDATE;
    IF FOUND THEN
        RETURN (v_existing.recovery_set_id,v_existing.environment_id,
                v_existing.plan_sha256,v_existing.manifest_sha256,
                v_existing.receipt_sha256,v_existing.receipt_bytes,
                v_existing.new_database_system_id,
                v_existing.switch_ack_sha256,
                v_existing.component_inventory)
            IS NOT DISTINCT FROM
               (p_recovery_set_id,p_environment_id,p_plan_sha256,
                p_manifest_sha256,p_receipt_sha256,p_receipt_bytes,
                p_new_database_system_id,p_switch_ack_sha256,
                p_component_inventory)
           AND EXISTS (
               SELECT 1 FROM platform.platformops_database_lineages
               WHERE environment_id=p_environment_id
                 AND current_database_system_id=p_new_database_system_id
           );
    END IF;
    INSERT INTO platform.platformops_restore_receipt_imports(
        operation_id,recovery_set_id,environment_id,plan_sha256,
        manifest_sha256,receipt_sha256,receipt_bytes,
        previous_database_system_id,new_database_system_id,
        object_merkle_sha256,locator_remap_sha256,safety_evidence_sha256,
        drain_ack_sha256,switch_ack_sha256,
        control_database_locator_sha256,scratch_database_locator_sha256,
        component_inventory,imported_at
    ) VALUES (
        p_operation_id,p_recovery_set_id,p_environment_id,p_plan_sha256,
        p_manifest_sha256,p_receipt_sha256,p_receipt_bytes,
        p_previous_database_system_id,p_new_database_system_id,
        p_object_merkle_sha256,p_locator_remap_sha256,p_safety_evidence_sha256,
        p_drain_ack_sha256,p_switch_ack_sha256,
        p_control_database_locator_sha256,p_scratch_database_locator_sha256,
        p_component_inventory,v_now
    );
    UPDATE platform.platformops_database_lineages
    SET current_database_system_id=p_new_database_system_id,
        generation=generation+1,updated_at=v_now
    WHERE environment_id=p_environment_id
      AND current_database_system_id=p_previous_database_system_id
    RETURNING generation INTO v_lineage_generation;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='restored database lineage import denied';
    END IF;
    INSERT INTO platform.platformops_database_lineage_events(
        operation_id,environment_id,previous_database_system_id,
        database_system_id,generation,event_kind,occurred_at
    ) VALUES (
        p_operation_id,p_environment_id,p_previous_database_system_id,
        p_new_database_system_id,v_lineage_generation,'restored',v_now
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_import_restore_receipt_v3(
    uuid,uuid,uuid,text,text,text,bytea,uuid,uuid,text,text,text,
    text,text,text,text,jsonb
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platformops_import_restore_receipt_v2(
    uuid,uuid,uuid,text,text,text,bytea,uuid,uuid,text,text,text,jsonb
) FROM harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_import_restore_receipt_v3(
    uuid,uuid,uuid,text,text,text,bytea,uuid,uuid,text,text,text,
    text,text,text,text,jsonb
) TO harpy_restore;
