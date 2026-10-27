-- +goose Up
CREATE TABLE platform.platformops_restore_receipt_imports (
    operation_id uuid PRIMARY KEY,
    recovery_set_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    receipt_sha256 text NOT NULL UNIQUE
        CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    receipt_bytes bytea NOT NULL
        CHECK (pg_catalog.octet_length(receipt_bytes) BETWEEN 1 AND 2097152),
    previous_database_system_id uuid,
    new_database_system_id uuid NOT NULL,
    object_merkle_sha256 text NOT NULL
        CHECK (object_merkle_sha256 ~ '^[0-9a-f]{64}$'),
    locator_remap_sha256 text NOT NULL
        CHECK (locator_remap_sha256 ~ '^[0-9a-f]{64}$'),
    safety_evidence_sha256 text NOT NULL
        CHECK (safety_evidence_sha256 ~ '^[0-9a-f]{64}$'),
    component_inventory jsonb NOT NULL,
    imported_at timestamptz NOT NULL
);

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_import_restore_receipt_v1(
    p_operation_id uuid,p_recovery_set_id uuid,p_environment_id uuid,
    p_plan_sha256 text,p_manifest_sha256 text,p_receipt_sha256 text,
    p_receipt_bytes bytea,p_previous_database_system_id uuid,
    p_new_database_system_id uuid,p_object_merkle_sha256 text,
    p_locator_remap_sha256 text,p_safety_evidence_sha256 text,
    p_component_inventory jsonb
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.statement_timestamp();
    v_receipt jsonb;
    v_existing platform.platformops_restore_receipt_imports%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-restore'
       OR p_operation_id IS NULL OR p_recovery_set_id IS NULL
       OR p_environment_id IS NULL OR p_new_database_system_id IS NULL
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$'
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$'
       OR p_receipt_sha256 !~ '^[0-9a-f]{64}$'
       OR p_object_merkle_sha256 !~ '^[0-9a-f]{64}$'
       OR p_locator_remap_sha256 !~ '^[0-9a-f]{64}$'
       OR p_safety_evidence_sha256 !~ '^[0-9a-f]{64}$'
       OR pg_catalog.octet_length(p_receipt_bytes) NOT BETWEEN 1 AND 2097152
       OR p_receipt_sha256<>pg_catalog.encode(
           pg_catalog.sha256(p_receipt_bytes),'hex'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore receipt import denied';
    END IF;
    IF pg_catalog.jsonb_typeof(p_component_inventory)<>'object'
       OR pg_catalog.jsonb_object_length(p_component_inventory)<>5
       OR NOT p_component_inventory ?& ARRAY[
           'database_archive','documents_inventory','key_custody',
           'owner_anchor','protected_authority'
       ] OR EXISTS (
           SELECT 1
           FROM pg_catalog.jsonb_each(p_component_inventory) AS component
           WHERE pg_catalog.jsonb_typeof(component.value)<>'object'
              OR component.value->>'object_key' IS NULL
              OR pg_catalog.length(component.value->>'object_key') NOT BETWEEN 1 AND 1024
              OR component.value->>'object_version_id' IS NULL
              OR pg_catalog.length(component.value->>'object_version_id') NOT BETWEEN 1 AND 1024
              OR component.value->>'ciphertext_sha256' !~ '^[0-9a-f]{64}$'
              OR (component.value->>'ciphertext_bytes')::bigint<=0
       ) THEN
        RAISE EXCEPTION USING ERRCODE='23514',
            MESSAGE='restore component import is invalid';
    END IF;
    v_receipt := pg_catalog.convert_from(p_receipt_bytes,'UTF8')::jsonb;
    IF v_receipt->>'schema' IS DISTINCT FROM 'forgecrud-recovery-artifact.v1'
       OR v_receipt->>'purpose' IS DISTINCT FROM 'restore_receipt'
       OR v_receipt->>'environment_id' IS DISTINCT FROM p_environment_id::text
       OR v_receipt->>'operation_id' IS DISTINCT FROM p_operation_id::text
       OR v_receipt->>'artifact_id' IS DISTINCT FROM p_operation_id::text
       OR v_receipt->>'algorithm' IS DISTINCT FROM 'Ed25519'
       OR v_receipt->>'payload_sha256' !~ '^[0-9a-f]{64}$'
       OR pg_catalog.octet_length(v_receipt->>'signature')<1
       OR (v_receipt->>'issued_at')::timestamptz>v_now
       OR (v_receipt->>'expires_at')::timestamptz<=v_now
       OR v_receipt#>>'{payload,schema}' IS DISTINCT FROM
          'restore-receipt-payload.v1'
       OR v_receipt#>>'{payload,recovery_set_id}' IS DISTINCT FROM
          p_recovery_set_id::text
       OR v_receipt#>>'{payload,manifest_sha256}' IS DISTINCT FROM
          p_manifest_sha256
       OR v_receipt#>>'{payload,identity,environment_id}' IS DISTINCT FROM
          p_environment_id::text
       OR v_receipt#>>'{payload,new_database_system_id}' IS DISTINCT FROM
          p_new_database_system_id::text
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
            MESSAGE='restore receipt evidence import denied';
    END IF;
    SELECT * INTO v_existing
    FROM platform.platformops_restore_receipt_imports
    WHERE operation_id=p_operation_id FOR UPDATE;
    IF FOUND THEN
        RETURN (v_existing.recovery_set_id,v_existing.environment_id,
                v_existing.plan_sha256,v_existing.manifest_sha256,
                v_existing.receipt_sha256,v_existing.receipt_bytes,
                v_existing.new_database_system_id,
                v_existing.component_inventory)
            IS NOT DISTINCT FROM
               (p_recovery_set_id,p_environment_id,p_plan_sha256,
                p_manifest_sha256,p_receipt_sha256,p_receipt_bytes,
                p_new_database_system_id,p_component_inventory);
    END IF;
    INSERT INTO platform.platformops_restore_receipt_imports(
        operation_id,recovery_set_id,environment_id,plan_sha256,
        manifest_sha256,receipt_sha256,receipt_bytes,
        previous_database_system_id,new_database_system_id,
        object_merkle_sha256,locator_remap_sha256,
        safety_evidence_sha256,component_inventory,imported_at
    ) VALUES (
        p_operation_id,p_recovery_set_id,p_environment_id,p_plan_sha256,
        p_manifest_sha256,p_receipt_sha256,p_receipt_bytes,
        p_previous_database_system_id,p_new_database_system_id,
        p_object_merkle_sha256,p_locator_remap_sha256,
        p_safety_evidence_sha256,p_component_inventory,v_now
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_import_restore_receipt_v1(
    uuid,uuid,uuid,text,text,text,bytea,uuid,uuid,text,text,text,jsonb
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_import_restore_receipt_v1(
    uuid,uuid,uuid,text,text,text,bytea,uuid,uuid,text,text,text,jsonb
) TO forgecrud_restore;
