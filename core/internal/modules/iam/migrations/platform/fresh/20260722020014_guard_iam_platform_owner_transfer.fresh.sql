-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_protected_authority_head_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Protected authority head deletion is forbidden';
    ELSIF TG_OP = 'INSERT' AND NEW.revision <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Protected authority head origin is invalid';
    ELSIF TG_OP = 'UPDATE' AND (
        NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.registry_binding_id IS DISTINCT FROM OLD.registry_binding_id OR
        NEW.scope_kind IS DISTINCT FROM OLD.scope_kind OR NEW.scope_id IS DISTINCT FROM OLD.scope_id OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR NEW.generation <= OLD.generation OR
        NEW.revision IS DISTINCT FROM OLD.revision + 1
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Protected authority head CAS is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.iam_owner_transfer_operation_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Owner transfer operation deletion is forbidden';
    ELSIF TG_OP = 'INSERT' AND (NEW.status <> 'prepared' OR NEW.revision <> 1) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Owner transfer operation origin is invalid';
    ELSIF TG_OP = 'UPDATE' AND (
        NEW.operation_id IS DISTINCT FROM OLD.operation_id OR
        NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.request_sha256 IS DISTINCT FROM OLD.request_sha256 OR
        NEW.actor_user_id IS DISTINCT FROM OLD.actor_user_id OR
        NEW.successor_user_id IS DISTINCT FROM OLD.successor_user_id OR
        NEW.actor_expected_revision IS DISTINCT FROM OLD.actor_expected_revision OR
        NEW.actor_expected_token_version IS DISTINCT FROM OLD.actor_expected_token_version OR
        NEW.actor_expected_authorization_version IS DISTINCT FROM OLD.actor_expected_authorization_version OR
        NEW.successor_expected_revision IS DISTINCT FROM OLD.successor_expected_revision OR
        NEW.successor_expected_token_version IS DISTINCT FROM OLD.successor_expected_token_version OR
        NEW.successor_expected_authorization_version IS DISTINCT FROM OLD.successor_expected_authorization_version OR
        NEW.successor_subject_id IS DISTINCT FROM OLD.successor_subject_id OR
        NEW.successor_binding_id IS DISTINCT FROM OLD.successor_binding_id OR
        NEW.expected_owner_policy_revision IS DISTINCT FROM OLD.expected_owner_policy_revision OR
        NEW.expected_platform_clock IS DISTINCT FROM OLD.expected_platform_clock OR
        NEW.registry_binding_id IS DISTINCT FROM OLD.registry_binding_id OR
        NEW.scope_kind IS DISTINCT FROM OLD.scope_kind OR NEW.scope_id IS DISTINCT FROM OLD.scope_id OR
        NEW.generation IS DISTINCT FROM OLD.generation OR
        NEW.previous_head_sha256 IS DISTINCT FROM OLD.previous_head_sha256 OR
        NEW.state_sha256 IS DISTINCT FROM OLD.state_sha256 OR
        NEW.proposal_sha256 IS DISTINCT FROM OLD.proposal_sha256 OR
        NEW.proposal_signer_kid IS DISTINCT FROM OLD.proposal_signer_kid OR
        NEW.proposal_signature_sha256 IS DISTINCT FROM OLD.proposal_signature_sha256 OR
        NEW.proposal_expires_at IS DISTINCT FROM OLD.proposal_expires_at OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR NEW.revision IS DISTINCT FROM OLD.revision + 1 OR
        NOT ((OLD.status = 'prepared' AND NEW.status = 'ack_verified') OR
             (OLD.status = 'ack_verified' AND NEW.status IN ('completed','compensated')))
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer operation transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.iam_owner_transfer_compensation_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Owner transfer compensation deletion is forbidden';
    ELSIF TG_OP = 'INSERT' AND (NEW.status <> 'prepared' OR NEW.revision <> 1) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Owner transfer compensation origin is invalid';
    ELSIF TG_OP = 'UPDATE' AND (
        NEW.operation_id IS DISTINCT FROM OLD.operation_id OR
        NEW.original_operation_id IS DISTINCT FROM OLD.original_operation_id OR
        NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.request_sha256 IS DISTINCT FROM OLD.request_sha256 OR
        NEW.actor_user_id IS DISTINCT FROM OLD.actor_user_id OR
        NEW.expected_platform_clock IS DISTINCT FROM OLD.expected_platform_clock OR
        NEW.registry_binding_id IS DISTINCT FROM OLD.registry_binding_id OR
        NEW.scope_kind IS DISTINCT FROM OLD.scope_kind OR NEW.scope_id IS DISTINCT FROM OLD.scope_id OR
        NEW.generation IS DISTINCT FROM OLD.generation OR
        NEW.previous_head_sha256 IS DISTINCT FROM OLD.previous_head_sha256 OR
        NEW.state_sha256 IS DISTINCT FROM OLD.state_sha256 OR
        NEW.proposal_sha256 IS DISTINCT FROM OLD.proposal_sha256 OR
        NEW.proposal_signer_kid IS DISTINCT FROM OLD.proposal_signer_kid OR
        NEW.proposal_signature_sha256 IS DISTINCT FROM OLD.proposal_signature_sha256 OR
        NEW.proposal_expires_at IS DISTINCT FROM OLD.proposal_expires_at OR
        NEW.compensates_ack_sha256 IS DISTINCT FROM OLD.compensates_ack_sha256 OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR NEW.revision IS DISTINCT FROM OLD.revision + 1 OR
        OLD.status <> 'prepared' OR NEW.status <> 'completed'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer compensation transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER iam_protected_authority_heads_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_protected_authority_heads
FOR EACH ROW EXECUTE FUNCTION platform.iam_protected_authority_head_guard_v1();
ALTER TABLE platform.iam_protected_authority_heads ENABLE ALWAYS TRIGGER iam_protected_authority_heads_guard;

CREATE TRIGGER iam_owner_transfer_operations_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_platform_owner_transfer_operations
FOR EACH ROW EXECUTE FUNCTION platform.iam_owner_transfer_operation_guard_v1();
ALTER TABLE platform.iam_platform_owner_transfer_operations ENABLE ALWAYS TRIGGER iam_owner_transfer_operations_guard;

CREATE TRIGGER iam_owner_transfer_compensations_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_platform_owner_transfer_compensations
FOR EACH ROW EXECUTE FUNCTION platform.iam_owner_transfer_compensation_guard_v1();
ALTER TABLE platform.iam_platform_owner_transfer_compensations ENABLE ALWAYS TRIGGER iam_owner_transfer_compensations_guard;

REVOKE ALL ON FUNCTION platform.iam_protected_authority_head_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_owner_transfer_operation_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_owner_transfer_compensation_guard_v1() FROM PUBLIC;
