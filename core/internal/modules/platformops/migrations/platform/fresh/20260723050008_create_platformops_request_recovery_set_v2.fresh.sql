-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_request_recovery_set_v2(
    p_recovery_set_id uuid,
    p_operation_id uuid,
    p_environment_id uuid,
    p_created_at timestamptz
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.platformops_recovery_sets%ROWTYPE;
    v_primary_binding_id uuid;
    v_tenant_set_sha256 text;
    v_owner_binding_id uuid;
    v_owner_binding_generation bigint;
    v_database_lineage_id uuid;
    v_signing_generation bigint;
    v_trust_bundle_sha256 text;
    v_sequence bigint;
BEGIN
    IF p_recovery_set_id IS NULL OR p_operation_id IS NULL
       OR p_environment_id IS NULL OR p_created_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'invalid recovery set request';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'platformops:recovery:' || p_environment_id::text,0
        )
    );
    SELECT * INTO v_existing
      FROM platform.platformops_recovery_sets
     WHERE operation_id = p_operation_id
        OR recovery_set_id = p_recovery_set_id
     ORDER BY operation_id = p_operation_id DESC
     LIMIT 1;
    IF FOUND THEN
        IF v_existing.recovery_set_id <> p_recovery_set_id
           OR v_existing.operation_id <> p_operation_id
           OR v_existing.environment_id <> p_environment_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'recovery set request conflicts';
        END IF;
        RETURN v_existing.sequence;
    END IF;
    SELECT identity.primary_binding_id,identity.tenant_set_sha256,
           identity.owner_binding_id,identity.owner_binding_generation,
           identity.database_lineage_id
      INTO v_primary_binding_id,v_tenant_set_sha256,
           v_owner_binding_id,v_owner_binding_generation,
           v_database_lineage_id
      FROM platform.platformops_current_environment_identity_v1(
          p_environment_id
      ) AS identity;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'current on-premise recovery identity is unavailable';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.platformops_recovery_protections AS protection
         WHERE protection.environment_id=p_environment_id
           AND protection.state='protected' AND protection.active_kit_sha256 IS NOT NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='acknowledged recovery protection is required';
    END IF;
    SELECT bundle.generation,bundle.bundle_sha256
      INTO v_signing_generation,v_trust_bundle_sha256
      FROM platform.platformops_recovery_signing_bundles AS bundle
     WHERE bundle.environment_id = p_environment_id
     ORDER BY bundle.generation DESC
     LIMIT 1;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'current recovery signing bundle is missing';
    END IF;
    SELECT COALESCE(max(recovery.sequence),0) + 1
      INTO v_sequence
      FROM platform.platformops_recovery_sets AS recovery
     WHERE recovery.environment_id = p_environment_id;
    INSERT INTO platform.platformops_recovery_sets(
        recovery_set_id,operation_id,environment_id,deployment_mode,
        portability,sequence,state,primary_binding_id,tenant_set_sha256,
        owner_binding_id,owner_binding_generation,database_lineage_id,
        recovery_signing_generation,recovery_trust_bundle_sha256,
        created_at,updated_at
    ) VALUES (
        p_recovery_set_id,p_operation_id,p_environment_id,'onpremise',
        'same_environment',v_sequence,'requested',v_primary_binding_id,
        v_tenant_set_sha256,v_owner_binding_id,v_owner_binding_generation,
        v_database_lineage_id,v_signing_generation,v_trust_bundle_sha256,
        p_created_at,p_created_at
    );
    RETURN v_sequence;
END;
$function$;
-- +goose StatementEnd

REVOKE EXECUTE ON FUNCTION platform.platformops_request_recovery_set_v1(
    uuid,uuid,uuid,uuid,text,uuid,bigint,uuid,bigint,text,timestamptz
) FROM forgecrud_runtime;
REVOKE ALL ON FUNCTION platform.platformops_request_recovery_set_v2(
    uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_request_recovery_set_v2(
    uuid,uuid,uuid,timestamptz
) TO forgecrud_runtime;
