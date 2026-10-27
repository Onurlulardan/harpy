-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.verify_finalization_receipt_v1(
    p_operation_id uuid, p_run_id uuid, p_finalizer_lease_id uuid,
    p_terminal_state text, p_terminal_evidence_sha256 text
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_receipt cutover_authority.finalization_receipts%ROWTYPE;
    v_authorization cutover_authority.finalization_authorizations%ROWTYPE;
    v_finalizer jsonb;
    v_catalog jsonb;
    v_current_generation integer;
    v_current_receipt_sha256 text;
    v_bound_generation integer;
    v_inventory_count bigint;
    v_valid_chain_count bigint;
BEGIN
    IF SESSION_USER <> 'harpy_cutover_finalizer' OR p_operation_id IS NULL OR
       p_run_id IS NULL OR p_finalizer_lease_id IS NULL OR
       p_terminal_state NOT IN ('completed','failed','cancelled') OR
       p_terminal_evidence_sha256 IS NULL OR p_terminal_evidence_sha256!~'^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='cutover finalization receipt verifier denied';
    END IF;
    SELECT receipt.* INTO STRICT v_receipt FROM cutover_authority.finalization_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id
      AND receipt.finalizer_lease_id=p_finalizer_lease_id
      AND receipt.terminal_state=p_terminal_state
      AND receipt.terminal_evidence_sha256=p_terminal_evidence_sha256;
    SELECT authorization_row.* INTO STRICT v_authorization
    FROM cutover_authority.finalization_authorizations AS authorization_row
    WHERE authorization_row.operation_id=p_operation_id AND authorization_row.run_id=p_run_id
      AND authorization_row.finalizer_lease_id=p_finalizer_lease_id;
    IF v_receipt.finalizer_lease_id IS DISTINCT FROM v_authorization.finalizer_lease_id OR
       v_receipt.finalizer_lease_sha256 IS DISTINCT FROM v_authorization.finalizer_lease_sha256 OR
       v_receipt.finalizer_lease_document IS DISTINCT FROM v_authorization.finalizer_lease_document OR
       v_receipt.adoption_lease_sha256 IS DISTINCT FROM v_authorization.adoption_lease_sha256 OR
       v_receipt.adoption_lease_document IS DISTINCT FROM v_authorization.adoption_lease_document OR
       v_receipt.seal_receipt_sha256 IS DISTINCT FROM v_authorization.seal_receipt_sha256 OR
       v_receipt.seal_receipt_document IS DISTINCT FROM v_authorization.seal_receipt_document OR
       v_receipt.cleanup_receipt_sha256 IS DISTINCT FROM v_authorization.cleanup_receipt_sha256 OR
       v_receipt.cleanup_receipt_document IS DISTINCT FROM v_authorization.cleanup_receipt_document OR
       v_receipt.authority_receipt_sha256 IS DISTINCT FROM v_authorization.authority_receipt_sha256 OR
       v_receipt.authority_receipt_document IS DISTINCT FROM v_authorization.authority_receipt_document OR
       v_receipt.terminal_state IS DISTINCT FROM v_authorization.terminal_state OR
       v_receipt.terminal_evidence_sha256 IS DISTINCT FROM v_authorization.terminal_evidence_sha256 OR
       v_receipt.terminal_evidence_document IS DISTINCT FROM v_authorization.terminal_evidence_document OR
       v_receipt.finalization_state_sha256 IS DISTINCT FROM v_authorization.finalization_state_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover protected finalization receipt differs';
    END IF;
    v_finalizer:=pg_catalog.convert_from(v_receipt.finalizer_lease_document,'UTF8')::jsonb;
    IF p_terminal_state='completed' THEN
        v_catalog:=pg_catalog.convert_from(v_receipt.tenant_requirement_catalog_document,'UTF8')::jsonb;
        IF v_finalizer->>'tenant_requirement_catalog_sha256' IS DISTINCT FROM v_receipt.tenant_requirement_catalog_sha256 OR
            (v_finalizer->>'tenant_requirement_count')::integer IS DISTINCT FROM v_receipt.tenant_requirement_count OR
            pg_catalog.encode(pg_catalog.sha256(v_receipt.tenant_requirement_catalog_document),'hex')
              IS DISTINCT FROM v_receipt.tenant_requirement_catalog_sha256 OR
            pg_catalog.jsonb_typeof(v_catalog) IS DISTINCT FROM 'object' OR
            v_catalog->>'contract_version' IS DISTINCT FROM 'tenant-migration-requirement-catalog.v2' OR
            pg_catalog.jsonb_typeof(v_catalog->'requirements') IS DISTINCT FROM 'array' OR
            pg_catalog.jsonb_array_length(v_catalog->'requirements') IS DISTINCT FROM v_receipt.tenant_requirement_count OR
            pg_catalog.jsonb_typeof(v_catalog->'targets') IS DISTINCT FROM 'array' OR
            pg_catalog.jsonb_array_length(v_catalog->'targets') IS DISTINCT FROM v_receipt.tenant_target_count THEN
            RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover tenant finalization receipt differs';
        END IF;
    ELSIF v_receipt.tenant_requirement_catalog_sha256 IS NOT NULL OR
          v_receipt.tenant_requirement_catalog_document IS NOT NULL OR
          v_receipt.tenant_requirement_count IS NOT NULL OR v_receipt.tenant_target_count IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover abort finalization receipt differs';
    END IF;
    SELECT inventory.catalog_generation,inventory.receipt_sha256
    INTO STRICT v_current_generation,v_current_receipt_sha256
    FROM cutover_authority.authority_adoption_inventory AS inventory
    ORDER BY inventory.catalog_generation DESC LIMIT 1;
    PERFORM cutover_authority.assert_adopted_catalog_v1(v_current_receipt_sha256);
    SELECT inventory.catalog_generation INTO v_bound_generation
    FROM cutover_authority.authority_adoption_inventory AS inventory
    WHERE inventory.receipt_sha256=v_receipt.authority_receipt_sha256
      AND inventory.receipt_document=v_receipt.authority_receipt_document
      AND pg_catalog.encode(pg_catalog.sha256(inventory.receipt_document),'hex')=
          inventory.receipt_sha256;
    WITH decoded AS (
        SELECT inventory.*,
            pg_catalog.convert_from(inventory.receipt_document,'UTF8')::jsonb AS payload
        FROM cutover_authority.authority_adoption_inventory AS inventory
    ), linked AS (
        SELECT current_row.*,previous.receipt_sha256 AS prior_sha256,
            previous.catalog_contract_sha256 AS prior_contract_sha256,
            previous.receipt_id AS prior_receipt_id,
            previous.environment_id AS prior_environment_id,
            previous.catalog_generation AS prior_generation,
            previous.catalog_instance_sha256 AS prior_instance_sha256,
            previous.payload AS prior_payload
        FROM decoded AS current_row LEFT JOIN decoded AS previous
          ON previous.catalog_generation=current_row.catalog_generation-1
    )
    SELECT pg_catalog.count(*),pg_catalog.count(*) FILTER (WHERE
        pg_catalog.encode(pg_catalog.sha256(receipt_document),'hex')=receipt_sha256 AND
        payload->>'contract_version'='cutover-authority-adoption-receipt.v1' AND
        payload->>'result'='succeeded' AND
        (payload->>'catalog_generation')::integer=catalog_generation AND
        payload->>'receipt_id'=receipt_id::text AND
        payload->>'operation_id'=operation_id::text AND
        payload->>'operation_sha256'=operation_sha256 AND
        payload->>'release_id'=release_id::text AND
        payload->>'environment_id'=environment_id::text AND
        payload->'target'->>'database_name'=pg_catalog.current_database() AND
        payload->'target'->>'database_origin'=database_origin AND
        payload->'target'->>'system_identifier'=system_identifier::text AND
        payload->'target'->>'database_oid'=database_oid::text AND
        (payload->>'adopted_at')::timestamptz=adopted_at AND
        payload->>'post_catalog_contract_sha256'=catalog_contract_sha256 AND
        payload->>'catalog_instance_sha256'=catalog_instance_sha256 AND
        payload->>'dependency_contract_sha256'=dependency_contract_sha256 AND
        payload->>'dependency_instance_sha256'=dependency_instance_sha256 AND
        pg_catalog.jsonb_array_length(payload->'objects')=object_count AND
        pg_catalog.jsonb_array_length(payload->'dependencies')=dependency_count AND
        CASE WHEN catalog_generation=1 THEN
            payload->>'mode'='initial_adoption' AND
            (payload->>'previous_catalog_generation')::integer=0 AND
            payload->>'pre_catalog_contract_sha256'=
                'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855' AND
            payload->'previous_authority_receipt_sha256'='null'::jsonb AND
            payload->'previous_catalog_checkpoint'='null'::jsonb
        ELSE payload->>'mode'='catalog_upgrade' AND
            (payload->>'previous_catalog_generation')::integer=catalog_generation-1 AND
            payload->>'previous_authority_receipt_sha256'=prior_sha256 AND
            payload->>'pre_catalog_contract_sha256'=prior_contract_sha256 AND
            (payload->>'adopted_at')::timestamptz>
                (prior_payload->>'adopted_at')::timestamptz AND
            pg_catalog.jsonb_array_length(payload->'objects')>
                pg_catalog.jsonb_array_length(prior_payload->'objects') AND
            NOT EXISTS (SELECT 1
                FROM pg_catalog.jsonb_array_elements(prior_payload->'objects')
                    WITH ORDINALITY AS prior_item(value,ordinal)
                LEFT JOIN pg_catalog.jsonb_array_elements(payload->'objects')
                    WITH ORDINALITY AS current_item(value,ordinal) USING (ordinal)
                WHERE current_item.value IS DISTINCT FROM prior_item.value) AND
            payload->'previous_catalog_checkpoint'=pg_catalog.jsonb_build_object(
                'receipt_sha256',prior_sha256,'receipt_id',prior_receipt_id,
                'environment_id',prior_environment_id,'target',prior_payload->'target',
                'catalog_generation',prior_generation,'adopted_at',prior_payload->>'adopted_at',
                'catalog_contract_sha256',prior_contract_sha256,
                'catalog_instance_sha256',prior_instance_sha256,
                'objects',prior_payload->'objects') END
    ) INTO v_inventory_count,v_valid_chain_count FROM linked;
    IF v_bound_generation IS NULL OR v_bound_generation>v_current_generation OR
       v_inventory_count IS DISTINCT FROM v_current_generation::bigint OR
       v_valid_chain_count IS DISTINCT FROM v_inventory_count THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover authority receipt chain differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.verify_finalization_receipt_v1(uuid,uuid,uuid,text,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION cutover_authority.verify_finalization_receipt_v1(uuid,uuid,uuid,text,text) TO harpy_cutover_finalizer;
