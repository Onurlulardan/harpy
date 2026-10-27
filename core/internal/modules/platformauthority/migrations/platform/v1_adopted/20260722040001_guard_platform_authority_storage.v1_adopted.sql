-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_reject_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority evidence is append only';
END;
$function$;

CREATE FUNCTION platform.platformauthority_guard_operation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority operation delete is forbidden';
    END IF;
    IF ROW(NEW.operation_id, NEW.binding_id, NEW.purpose, NEW.scope_kind, NEW.scope_id,
        NEW.generation, NEW.previous_head_sha256, NEW.state_sha256, NEW.state_bytes,
        NEW.proposal_sha256, NEW.proposal_bytes, NEW.compensates_ack_sha256,
        NEW.issued_at, NEW.expires_at, NEW.created_at)
       IS DISTINCT FROM
       ROW(OLD.operation_id, OLD.binding_id, OLD.purpose, OLD.scope_kind, OLD.scope_id,
        OLD.generation, OLD.previous_head_sha256, OLD.state_sha256, OLD.state_bytes,
        OLD.proposal_sha256, OLD.proposal_bytes, OLD.compensates_ack_sha256,
        OLD.issued_at, OLD.expires_at, OLD.created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority plan is immutable';
    END IF;
    IF NEW.revision <> OLD.revision + 1 OR NEW.updated_at < OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'protected authority revision is stale';
    END IF;
    IF NOT ((OLD.status = 'proposal_signed' AND NEW.status = 'ack_verified') OR
            (OLD.status = 'ack_verified' AND NEW.status = 'completed')) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.platformauthority_guard_head_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority head delete is forbidden';
    END IF;
    IF ROW(NEW.binding_id, NEW.scope_kind, NEW.scope_id) IS DISTINCT FROM
       ROW(OLD.binding_id, OLD.scope_kind, OLD.scope_id) OR
       NEW.generation <> OLD.generation + 1 OR
       NEW.previous_head_sha256 IS DISTINCT FROM OLD.ack_sha256 OR
       NEW.updated_at < OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'protected authority head CAS is stale';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.platformauthority_guard_binding_v2()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' OR ROW(NEW.binding_id, NEW.environment_id, NEW.deployment_mode,
        NEW.registry_kind, NEW.license_binding_id, NEW.finalization_receipt_sha256,
        NEW.proof_sha256, NEW.proof_bytes, NEW.created_at)
        IS DISTINCT FROM ROW(OLD.binding_id, OLD.environment_id, OLD.deployment_mode,
        OLD.registry_kind, OLD.license_binding_id, OLD.finalization_receipt_sha256,
        OLD.proof_sha256, OLD.proof_bytes, OLD.created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority binding identity is immutable';
    END IF;
    IF NOT (
        (OLD.signer_status = 'active' AND NEW.signer_status = 'rekey_required'
            AND NEW.signer_generation = OLD.signer_generation
            AND NEW.signer_kid = OLD.signer_kid
            AND NEW.signer_public_key_sha256 = OLD.signer_public_key_sha256
            AND NEW.recovery_trust_bundle_sha256 IS NOT DISTINCT FROM
                OLD.recovery_trust_bundle_sha256)
        OR (NEW.signer_status = 'active' AND NEW.signer_generation = OLD.signer_generation + 1
            AND NEW.signer_kid <> OLD.signer_kid
            AND NEW.signer_public_key_sha256 <> OLD.signer_public_key_sha256
            AND ((NEW.deployment_mode = 'onpremise' AND
                    NEW.recovery_trust_bundle_sha256 <> OLD.recovery_trust_bundle_sha256)
                OR (NEW.deployment_mode = 'saas' AND
                    NEW.recovery_trust_bundle_sha256 IS NULL)))
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority signer transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER platformauthority_bindings_guard
BEFORE UPDATE OR DELETE ON platform.platformauthority_bindings
FOR EACH ROW EXECUTE FUNCTION platform.platformauthority_guard_binding_v2();

CREATE TRIGGER platformauthority_operations_guard
BEFORE UPDATE OR DELETE ON platform.platformauthority_operations
FOR EACH ROW EXECUTE FUNCTION platform.platformauthority_guard_operation_v1();

CREATE TRIGGER platformauthority_heads_guard
BEFORE UPDATE OR DELETE ON platform.platformauthority_heads
FOR EACH ROW EXECUTE FUNCTION platform.platformauthority_guard_head_v1();

CREATE TRIGGER platformauthority_receipts_guard
BEFORE UPDATE OR DELETE ON platform.platformauthority_receipts
FOR EACH ROW EXECUTE FUNCTION platform.platformauthority_reject_mutation_v1();
