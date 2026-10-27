-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_publish_recovery_kit_v1(
    p_operation_id uuid,p_artifact_bytes bytea,p_artifact_sha256 bytea,
    p_ack_verifier_sha256 bytea,p_expires_at timestamptz,p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_operation_id IS NULL OR pg_catalog.octet_length(p_artifact_bytes)=0 OR
       pg_catalog.sha256(p_artifact_bytes) IS DISTINCT FROM p_artifact_sha256 OR
       pg_catalog.octet_length(p_ack_verifier_sha256)<>32 OR
       p_expires_at<=p_occurred_at THEN
        RETURN false;
    END IF;
    UPDATE platform.platformops_recovery_control_operations
       SET artifact_bytes=p_artifact_bytes,artifact_sha256=p_artifact_sha256,
           ack_verifier_sha256=p_ack_verifier_sha256,expires_at=p_expires_at,
           state='kit_ready',kit_claim_state='ready',updated_at=p_occurred_at
     WHERE operation_id=p_operation_id AND operation_kind IN ('enable','rotate')
       AND state IN ('accepted','handoff_ready') AND kit_claim_state='pending';
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_claim_recovery_kit_v1(
    p_operation_id uuid,p_actor_user_id uuid,p_session_id uuid,p_occurred_at timestamptz
)
RETURNS TABLE(artifact_bytes bytea,artifact_sha256 text,expires_at timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.platformops_recovery_control_operations%ROWTYPE;
    v_tenant_id uuid;
BEGIN
    SELECT * INTO v_operation
      FROM platform.platformops_recovery_control_operations AS operation
     WHERE operation.operation_id=p_operation_id FOR UPDATE;
    SELECT identity.tenant_id INTO v_tenant_id
      FROM platform.platformops_installation_recovery_identities AS identity
     WHERE identity.environment_id=v_operation.environment_id
       AND identity.identity_state='active' FOR SHARE;
    IF NOT FOUND OR p_actor_user_id IS DISTINCT FROM v_operation.actor_user_id OR
       p_session_id IS NULL OR v_operation.state<>'kit_ready' OR
       v_operation.kit_claim_state NOT IN ('ready','downloaded') OR
       v_operation.expires_at<=p_occurred_at OR
       (v_operation.claimed_session_id IS NOT NULL AND
        v_operation.claimed_session_id IS DISTINCT FROM p_session_id) OR
       NOT platform.iam_platform_current_owner_v1(v_tenant_id,p_actor_user_id) THEN
        RETURN;
    END IF;
    UPDATE platform.platformops_recovery_control_operations
       SET claimed_session_id=p_session_id,kit_claim_state='downloaded',
           updated_at=p_occurred_at WHERE operation_id=p_operation_id;
    RETURN QUERY SELECT v_operation.artifact_bytes,
        pg_catalog.encode(v_operation.artifact_sha256,'hex'),v_operation.expires_at;
END;
$function$;

CREATE FUNCTION platform.platformops_ack_recovery_kit_v1(
    p_operation_id uuid,p_actor_user_id uuid,p_session_id uuid,
    p_artifact_sha256 bytea,p_ack_proof bytea,p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.platformops_recovery_control_operations%ROWTYPE;
    v_tenant_id uuid;
BEGIN
    SELECT * INTO v_operation
      FROM platform.platformops_recovery_control_operations AS operation
     WHERE operation.operation_id=p_operation_id FOR UPDATE;
    SELECT identity.tenant_id INTO v_tenant_id
      FROM platform.platformops_installation_recovery_identities AS identity
     WHERE identity.environment_id=v_operation.environment_id
       AND identity.identity_state='active' FOR SHARE;
    IF NOT FOUND OR v_operation.actor_user_id IS DISTINCT FROM p_actor_user_id OR
       v_operation.claimed_session_id IS DISTINCT FROM p_session_id OR
       v_operation.state<>'kit_ready' OR v_operation.kit_claim_state<>'downloaded' OR
       v_operation.expires_at<=p_occurred_at OR
       v_operation.artifact_sha256 IS DISTINCT FROM p_artifact_sha256 OR
       pg_catalog.sha256(p_ack_proof) IS DISTINCT FROM v_operation.ack_verifier_sha256 OR
       NOT platform.iam_platform_current_owner_v1(v_tenant_id,p_actor_user_id) THEN
        RETURN false;
    END IF;
    UPDATE platform.platformops_recovery_control_operations
       SET state='completed',kit_claim_state='acknowledged',artifact_bytes=NULL,
           ack_verifier_sha256=NULL,
           completed_at=p_occurred_at,updated_at=p_occurred_at
     WHERE operation_id=p_operation_id;
    UPDATE platform.platformops_recovery_protections
       SET state='protected',generation=generation+1,
           active_kit_sha256=p_artifact_sha256,current_operation_id=NULL,
           updated_at=p_occurred_at
     WHERE environment_id=v_operation.environment_id
       AND current_operation_id=p_operation_id;
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_complete_recovery_disable_v1(
    p_operation_id uuid,p_handoff_receipt_sha256 bytea,p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_environment_id uuid;
BEGIN
    IF pg_catalog.octet_length(p_handoff_receipt_sha256)<>32 THEN RETURN false; END IF;
    UPDATE platform.platformops_recovery_control_operations
       SET state='completed',completed_at=p_occurred_at,updated_at=p_occurred_at
     WHERE operation_id=p_operation_id AND operation_kind='disable' AND state='accepted'
    RETURNING environment_id INTO v_environment_id;
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE platform.platformops_recovery_protections
       SET state='unprotected',active_kit_sha256=NULL,current_operation_id=NULL,
           updated_at=p_occurred_at
     WHERE environment_id=v_environment_id AND current_operation_id=p_operation_id;
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_publish_recovery_kit_v1(
    uuid,bytea,bytea,bytea,timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_claim_recovery_kit_v1(
    uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_ack_recovery_kit_v1(
    uuid,uuid,uuid,bytea,bytea,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_complete_recovery_disable_v1(
    uuid,bytea,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_recovery_kit_v1(
    uuid,uuid,uuid,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_ack_recovery_kit_v1(
    uuid,uuid,uuid,bytea,bytea,timestamptz
) TO harpy_runtime;
