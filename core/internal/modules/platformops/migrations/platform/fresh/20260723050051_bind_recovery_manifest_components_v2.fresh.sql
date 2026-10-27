-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_complete_recovery_execution_v2(
    p_recovery_set_id uuid,p_attempt_id uuid,p_lease_token uuid,
    p_manifest_sha256 text,p_manifest_bytes bytea,p_database_system_id uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_inventory jsonb;
    v_manifest_inventory jsonb;
    v_result boolean;
BEGIN
    SELECT pg_catalog.jsonb_object_agg(
        component_kind,
        pg_catalog.jsonb_build_object(
            'object_key',object_key,
            'object_version_id',object_version_id,
            'ciphertext_sha256',ciphertext_sha256,
            'ciphertext_bytes',ciphertext_bytes
        )
    ) INTO v_inventory
    FROM platform.platformops_recovery_components
    WHERE recovery_set_id=p_recovery_set_id;
    v_manifest_inventory :=
        pg_catalog.convert_from(p_manifest_bytes,'UTF8')::jsonb
        #> '{payload,components}';
    IF pg_catalog.jsonb_typeof(v_manifest_inventory)<>'object'
       OR pg_catalog.jsonb_object_length(v_manifest_inventory)<>5
       OR v_manifest_inventory IS DISTINCT FROM v_inventory THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery manifest component binding denied';
    END IF;
    SELECT platform.platformops_complete_recovery_execution_v1(
        p_recovery_set_id,p_attempt_id,p_lease_token,p_manifest_sha256,
        p_manifest_bytes,p_database_system_id
    ) INTO v_result;
    RETURN v_result;
END;
$function$;

CREATE FUNCTION platform.platformops_claim_next_recovery_set_v3(
    p_attempt_id uuid,p_lease_token uuid
) RETURNS TABLE(
    attempt_id uuid,lease_token uuid,attempt_generation bigint,
    recovery_state text,recovery_set_id uuid,operation_id uuid,
    environment_id uuid,sequence bigint,primary_binding_id uuid,
    tenant_set_sha256 text,owner_binding_id uuid,
    owner_binding_generation bigint,database_lineage_id uuid,
    database_system_id uuid,recovery_signing_generation bigint,
    recovery_trust_bundle_sha256 text,component_inventory jsonb,
    database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
RETURN QUERY
WITH claimed AS MATERIALIZED (
    SELECT * FROM platform.platformops_claim_next_recovery_set_v2(
        p_attempt_id,p_lease_token
    )
)
SELECT claimed.attempt_id,claimed.lease_token,claimed.attempt_generation,
       claimed.recovery_state,claimed.recovery_set_id,claimed.operation_id,
       claimed.environment_id,claimed.sequence,claimed.primary_binding_id,
       claimed.tenant_set_sha256,claimed.owner_binding_id,
       claimed.owner_binding_generation,claimed.database_lineage_id,
       claimed.database_system_id,claimed.recovery_signing_generation,
       claimed.recovery_trust_bundle_sha256,
       COALESCE((
           SELECT pg_catalog.jsonb_object_agg(
               component_kind,
               pg_catalog.jsonb_build_object(
                   'object_key',object_key,
                   'object_version_id',object_version_id,
                   'ciphertext_sha256',ciphertext_sha256,
                   'ciphertext_bytes',ciphertext_bytes
               )
           )
           FROM platform.platformops_recovery_components
           WHERE recovery_set_id=claimed.recovery_set_id
       ),'{}'::jsonb),
       claimed.database_now
FROM claimed;
END;
$function$;

CREATE FUNCTION platform.platformops_claim_restore_execution_v3(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,p_plan_sha256 text
) RETURNS TABLE(
    attempt_id uuid,lease_token uuid,attempt_generation bigint,
    restore_state text,operation_id uuid,recovery_set_id uuid,
    environment_id uuid,target_disposition text,plan_sha256 text,
    plan_bytes bytea,safety_recovery_set_id uuid,
    safety_recovery_set_sha256 text,waiver_reason text,probe_sha256 text,
    destructive_confirmed boolean,manifest_sha256 text,manifest_bytes bytea,
    previous_database_system_id uuid,component_inventory jsonb,
    database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    WITH claimed AS MATERIALIZED (
        SELECT * FROM platform.platformops_claim_restore_execution_v2(
            p_operation_id,p_attempt_id,p_lease_token,p_plan_sha256
        )
    )
    SELECT claimed.*
    FROM claimed
    WHERE pg_catalog.convert_from(claimed.manifest_bytes,'UTF8')::jsonb
          #> '{payload,components}'
          IS NOT DISTINCT FROM claimed.component_inventory;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore manifest component binding denied';
    END IF;
END;
$function$;

CREATE FUNCTION platform.platformops_import_restore_receipt_v2(
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
    v_manifest_inventory jsonb;
    v_result boolean;
BEGIN
    SELECT pg_catalog.convert_from(manifest_bytes,'UTF8')::jsonb
           #> '{payload,components}'
    INTO v_manifest_inventory
    FROM platform.platformops_recovery_sets
    WHERE recovery_set_id=p_recovery_set_id
      AND environment_id=p_environment_id
      AND manifest_sha256=p_manifest_sha256
      AND state='protected';
    IF NOT FOUND OR v_manifest_inventory IS DISTINCT FROM p_component_inventory THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore receipt component binding denied';
    END IF;
    SELECT platform.platformops_import_restore_receipt_v1(
        p_operation_id,p_recovery_set_id,p_environment_id,p_plan_sha256,
        p_manifest_sha256,p_receipt_sha256,p_receipt_bytes,
        p_previous_database_system_id,p_new_database_system_id,
        p_object_merkle_sha256,p_locator_remap_sha256,
        p_safety_evidence_sha256,p_component_inventory
    ) INTO v_result;
    RETURN v_result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_complete_recovery_execution_v2(
    uuid,uuid,uuid,text,bytea,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_claim_next_recovery_set_v3(
    uuid,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_claim_restore_execution_v3(
    uuid,uuid,uuid,text
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_import_restore_receipt_v2(
    uuid,uuid,uuid,text,text,text,bytea,uuid,uuid,text,text,text,jsonb
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platformops_complete_recovery_execution_v1(
    uuid,uuid,uuid,text,bytea,uuid
) FROM harpy_backup_reader;
REVOKE EXECUTE ON FUNCTION platform.platformops_claim_next_recovery_set_v2(
    uuid,uuid
) FROM harpy_backup_reader;
REVOKE EXECUTE ON FUNCTION platform.platformops_claim_restore_execution_v2(
    uuid,uuid,uuid,text
) FROM harpy_restore;
REVOKE EXECUTE ON FUNCTION platform.platformops_import_restore_receipt_v1(
    uuid,uuid,uuid,text,text,text,bytea,uuid,uuid,text,text,text,jsonb
) FROM harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_complete_recovery_execution_v2(
    uuid,uuid,uuid,text,bytea,uuid
) TO harpy_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_next_recovery_set_v3(
    uuid,uuid
) TO harpy_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_restore_execution_v3(
    uuid,uuid,uuid,text
) TO harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_import_restore_receipt_v2(
    uuid,uuid,uuid,text,text,text,bytea,uuid,uuid,text,text,text,jsonb
) TO harpy_restore;
