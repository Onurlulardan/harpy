-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_request_recovery_control_v1(
    p_operation_id uuid,p_environment_id uuid,p_actor_user_id uuid,
    p_idempotency_key text,p_request_sha256 bytea,p_operation_kind text,
    p_claim_algorithm text,p_claim_public_key bytea,p_occurred_at timestamptz
)
RETURNS TABLE(operation_id uuid,state text,kit_claim_state text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_protection platform.platformops_recovery_protections%ROWTYPE;
    v_existing platform.platformops_recovery_control_operations%ROWTYPE;
    v_tenant_id uuid;
BEGIN
    IF p_operation_id IS NULL OR p_environment_id IS NULL OR p_actor_user_id IS NULL OR
       p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$' OR
       pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_operation_kind NOT IN ('enable','rotate','disable') OR p_occurred_at IS NULL OR
       (p_operation_kind IN ('enable','rotate') AND
        (p_claim_algorithm IS DISTINCT FROM
         'HPKE-v1-X25519-HKDF-SHA256-AES-256-GCM' OR
         pg_catalog.octet_length(p_claim_public_key) <> 32)) OR
       (p_operation_kind = 'disable' AND
        (p_claim_algorithm IS NOT NULL OR p_claim_public_key IS NOT NULL)) THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='recovery control request is invalid';
    END IF;
    SELECT * INTO v_existing
      FROM platform.platformops_recovery_control_operations AS operation
     WHERE operation.operation_id=p_operation_id FOR SHARE;
    IF FOUND THEN
        IF v_existing.environment_id IS DISTINCT FROM p_environment_id OR
           v_existing.actor_user_id IS DISTINCT FROM p_actor_user_id OR
           v_existing.idempotency_key IS DISTINCT FROM p_idempotency_key OR
           v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 OR
           v_existing.operation_kind IS DISTINCT FROM p_operation_kind OR
           v_existing.claim_algorithm IS DISTINCT FROM p_claim_algorithm OR
           v_existing.claim_public_key IS DISTINCT FROM p_claim_public_key THEN
            RAISE EXCEPTION USING ERRCODE='23505',MESSAGE='recovery control replay conflicts';
        END IF;
        RETURN QUERY SELECT v_existing.operation_id,v_existing.state,
                            v_existing.kit_claim_state;
        RETURN;
    END IF;
    SELECT identity.tenant_id INTO v_tenant_id
      FROM platform.platformops_installation_recovery_identities AS identity
     WHERE identity.environment_id=p_environment_id
       AND identity.deployment_mode='onpremise' AND identity.identity_state='active'
     FOR SHARE;
    IF v_tenant_id IS NULL OR NOT platform.iam_platform_current_owner_v1(
        v_tenant_id,p_actor_user_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='current owner is required';
    END IF;
    SELECT * INTO v_protection
      FROM platform.platformops_recovery_protections AS protection
     WHERE protection.environment_id=p_environment_id FOR UPDATE;
    IF NOT FOUND OR v_protection.current_operation_id IS NOT NULL OR
       (p_operation_kind='enable' AND v_protection.state<>'unprotected') OR
       (p_operation_kind IN ('rotate','disable') AND v_protection.state<>'protected') THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery protection state conflicts';
    END IF;
    IF p_operation_kind='disable' AND EXISTS (
        SELECT 1 FROM platform.platformops_recovery_sets AS recovery_set
         WHERE recovery_set.environment_id=p_environment_id
           AND recovery_set.state IN ('requested','holding','snapshotting','copying',
                                      'verifying','protected','deleting')
    ) THEN
        RAISE EXCEPTION USING ERRCODE='23000',MESSAGE='retained recovery set blocks disable';
    END IF;
    INSERT INTO platform.platformops_recovery_control_operations(
        operation_id,environment_id,operation_kind,actor_user_id,idempotency_key,
        request_sha256,claim_algorithm,claim_public_key,state,kit_claim_state,
        created_at,updated_at
    ) VALUES (
        p_operation_id,p_environment_id,p_operation_kind,p_actor_user_id,p_idempotency_key,
        p_request_sha256,p_claim_algorithm,p_claim_public_key,'accepted',
        CASE WHEN p_operation_kind='disable' THEN 'not_applicable' ELSE 'pending' END,
        p_occurred_at,p_occurred_at
    );
    UPDATE platform.platformops_recovery_protections
       SET current_operation_id=p_operation_id,
           state=CASE p_operation_kind WHEN 'rotate' THEN 'rotation_pending'
                 WHEN 'disable' THEN 'disable_pending' ELSE 'unprotected' END,
           updated_at=p_occurred_at
     WHERE environment_id=p_environment_id;
    RETURN QUERY SELECT p_operation_id,'accepted'::text,
        CASE WHEN p_operation_kind='disable' THEN 'not_applicable' ELSE 'pending' END;
END;
$function$;

CREATE FUNCTION platform.platformops_get_recovery_protection_v1(p_environment_id uuid)
RETURNS TABLE(environment_id uuid,state text,generation bigint,
              kit_acknowledged boolean,updated_at timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT protection.environment_id,protection.state,protection.generation,
           protection.active_kit_sha256 IS NOT NULL,protection.updated_at
    FROM platform.platformops_recovery_protections AS protection
    WHERE protection.environment_id=p_environment_id;
END;
$function$;

CREATE FUNCTION platform.platformops_get_recovery_control_operation_v1(p_operation_id uuid)
RETURNS TABLE(operation_id uuid,environment_id uuid,operation_kind text,state text,
              kit_claim_state text,artifact_sha256 text,expires_at timestamptz,
              failure_code text,created_at timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT operation.operation_id,operation.environment_id,operation.operation_kind,
           operation.state,operation.kit_claim_state,
           pg_catalog.encode(operation.artifact_sha256,'hex'),operation.expires_at,
           operation.failure_code,operation.created_at
    FROM platform.platformops_recovery_control_operations AS operation
    WHERE operation.operation_id=p_operation_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_request_recovery_control_v1(
    uuid,uuid,uuid,text,bytea,text,text,bytea,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_get_recovery_protection_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_get_recovery_control_operation_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_request_recovery_control_v1(
    uuid,uuid,uuid,text,bytea,text,text,bytea,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_get_recovery_protection_v1(uuid)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_get_recovery_control_operation_v1(uuid)
TO harpy_runtime;
