-- +goose Up
ALTER TABLE platform.platformops_recovery_control_operations
DROP CONSTRAINT platformops_recovery_control_operations_operation_kind_check;
ALTER TABLE platform.platformops_recovery_control_operations
ADD CONSTRAINT platformops_recovery_control_operation_kind_ck CHECK (
    operation_kind IN ('enable','rotate','disable','anchor_register','anchor_delete')
);
ALTER TABLE platform.platformops_recovery_control_operations
DROP CONSTRAINT platformops_recovery_control_claim_ck;
ALTER TABLE platform.platformops_recovery_control_operations
ADD CONSTRAINT platformops_recovery_control_claim_ck CHECK (
    (operation_kind IN ('disable','anchor_register','anchor_delete')
        AND claim_algorithm IS NULL AND claim_public_key IS NULL
        AND kit_claim_state = 'not_applicable') OR
    (operation_kind IN ('enable','rotate')
        AND claim_algorithm = 'HPKE-v1-X25519-HKDF-SHA256-AES-256-GCM'
        AND pg_catalog.octet_length(claim_public_key) = 32
        AND kit_claim_state <> 'not_applicable')
);

ALTER TABLE platform.platformops_recovery_sets
DROP CONSTRAINT platformops_recovery_sets_state_check;
ALTER TABLE platform.platformops_recovery_sets
ADD CONSTRAINT platformops_recovery_sets_state_ck CHECK (state IN (
    'requested','holding','snapshotting','copying','verifying',
    'protected','stale','failed','deleting','deleted'
));
ALTER TABLE platform.platformops_recovery_sets
DROP CONSTRAINT platformops_recovery_sets_check1;
ALTER TABLE platform.platformops_recovery_sets
ADD CONSTRAINT platformops_recovery_sets_protected_at_ck CHECK (
    (state IN ('protected','stale')) = (protected_at IS NOT NULL)
);

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_record_anchor_operation_v1(
    p_operation_id uuid,p_environment_id uuid,p_actor_user_id uuid,
    p_idempotency_key text,p_request_sha256 bytea,p_kind text,p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_existing platform.platformops_recovery_control_operations%ROWTYPE;
BEGIN
    IF p_operation_id IS NULL OR p_environment_id IS NULL OR p_actor_user_id IS NULL OR
       p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$' OR
       pg_catalog.octet_length(p_request_sha256)<>32 OR
       p_kind NOT IN ('anchor_register','anchor_delete') OR p_occurred_at IS NULL THEN
        RETURN false;
    END IF;
    SELECT * INTO v_existing FROM platform.platformops_recovery_control_operations
     WHERE operation_id=p_operation_id FOR SHARE;
    IF FOUND THEN
        RETURN v_existing.environment_id=p_environment_id AND
            v_existing.actor_user_id=p_actor_user_id AND
            v_existing.idempotency_key=p_idempotency_key AND
            v_existing.request_sha256=p_request_sha256 AND
            v_existing.operation_kind=p_kind;
    END IF;
    INSERT INTO platform.platformops_recovery_control_operations(
        operation_id,environment_id,operation_kind,actor_user_id,idempotency_key,
        request_sha256,state,kit_claim_state,created_at,updated_at
    ) VALUES (p_operation_id,p_environment_id,p_kind,p_actor_user_id,
        p_idempotency_key,p_request_sha256,'accepted','not_applicable',
        p_occurred_at,p_occurred_at);
    RETURN true;
END;
$function$;

CREATE FUNCTION platform.platformops_mark_anchor_handoff_v1(
    p_operation_id uuid,p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    UPDATE platform.platformops_recovery_control_operations
       SET state='handoff_ready',updated_at=p_occurred_at
     WHERE operation_id=p_operation_id
       AND operation_kind IN ('anchor_register','anchor_delete')
       AND state IN ('accepted','handoff_ready');
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_complete_anchor_operation_v1(
    p_operation_id uuid,p_ack_sha256 bytea,p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF pg_catalog.octet_length(p_ack_sha256)<>32 THEN RETURN false; END IF;
    UPDATE platform.platformops_recovery_control_operations
       SET state='completed',completed_at=COALESCE(completed_at,p_occurred_at),
           updated_at=p_occurred_at
     WHERE operation_id=p_operation_id
       AND operation_kind IN ('anchor_register','anchor_delete')
       AND state IN ('accepted','handoff_ready','completed');
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_stale_recovery_sets_for_authority_v1(
    p_environment_id uuid,p_authority_generation bigint,
    p_state_sha256 bytea,p_occurred_at timestamptz
)
RETURNS bigint LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_count bigint;
BEGIN
    IF p_environment_id IS NULL OR p_authority_generation<1 OR
       pg_catalog.octet_length(p_state_sha256)<>32 OR p_occurred_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='authority stale request is invalid';
    END IF;
    UPDATE platform.platformops_recovery_sets
       SET state='stale',failure_code='protected_authority_advanced',updated_at=p_occurred_at
     WHERE environment_id=p_environment_id AND state='protected';
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;

CREATE FUNCTION platform.platformops_guard_recovery_set_v2()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NEW.recovery_set_id<>OLD.recovery_set_id OR NEW.operation_id<>OLD.operation_id OR
       NEW.environment_id<>OLD.environment_id OR NEW.deployment_mode<>OLD.deployment_mode OR
       NEW.portability<>OLD.portability OR NEW.sequence<>OLD.sequence OR
       NEW.primary_binding_id<>OLD.primary_binding_id OR NEW.tenant_set_sha256<>OLD.tenant_set_sha256 OR
       NEW.owner_binding_id<>OLD.owner_binding_id OR
       NEW.owner_binding_generation<>OLD.owner_binding_generation OR
       NEW.database_lineage_id<>OLD.database_lineage_id OR
       NEW.recovery_signing_generation<>OLD.recovery_signing_generation OR
       NEW.recovery_trust_bundle_sha256<>OLD.recovery_trust_bundle_sha256 OR
       NEW.created_at<>OLD.created_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',MESSAGE='recovery set identity is immutable';
    END IF;
    IF OLD.state<>NEW.state AND NOT (
        (OLD.state='requested' AND NEW.state IN ('holding','failed')) OR
        (OLD.state='holding' AND NEW.state IN ('snapshotting','failed')) OR
        (OLD.state='snapshotting' AND NEW.state IN ('copying','failed')) OR
        (OLD.state='copying' AND NEW.state IN ('verifying','failed')) OR
        (OLD.state='verifying' AND NEW.state IN ('protected','failed')) OR
        (OLD.state='protected' AND NEW.state IN ('stale','deleting')) OR
        (OLD.state='deleting' AND NEW.state IN ('protected','deleted'))
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',MESSAGE='recovery set transition is invalid';
    END IF;
    IF OLD.state='protected' AND NEW.state NOT IN ('stale','deleting') AND (
        NEW.manifest_sha256 IS DISTINCT FROM OLD.manifest_sha256 OR
        NEW.manifest_bytes IS DISTINCT FROM OLD.manifest_bytes OR
        NEW.database_system_id IS DISTINCT FROM OLD.database_system_id OR
        NEW.protected_at IS DISTINCT FROM OLD.protected_at) THEN
        RAISE EXCEPTION USING ERRCODE='55000',MESSAGE='protected recovery evidence is immutable';
    END IF;
    RETURN NEW;
END;
$function$;
CREATE OR REPLACE TRIGGER platformops_guard_recovery_set
BEFORE UPDATE ON platform.platformops_recovery_sets
FOR EACH ROW EXECUTE FUNCTION platform.platformops_guard_recovery_set_v2();
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_guard_recovery_set_v2() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_record_anchor_operation_v1(
    uuid,uuid,uuid,text,bytea,text,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_mark_anchor_handoff_v1(uuid,timestamptz)
FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_complete_anchor_operation_v1(
    uuid,bytea,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_stale_recovery_sets_for_authority_v1(
    uuid,bigint,bytea,timestamptz) FROM PUBLIC;
