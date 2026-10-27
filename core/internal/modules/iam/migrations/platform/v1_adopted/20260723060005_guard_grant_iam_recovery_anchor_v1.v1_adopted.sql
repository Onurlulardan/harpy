-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='recovery anchor deletion is forbidden';
    ELSIF TG_OP='INSERT' AND (NEW.state<>'pending' OR NEW.revision<>1) THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='recovery anchor origin is invalid';
    ELSIF TG_OP='UPDATE' AND (
        NEW.anchor_id IS DISTINCT FROM OLD.anchor_id OR
        NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.tenant_id IS DISTINCT FROM OLD.tenant_id OR
        NEW.owner_user_id IS DISTINCT FROM OLD.owner_user_id OR
        NEW.generation IS DISTINCT FROM OLD.generation OR
        NEW.credential_id IS DISTINCT FROM OLD.credential_id OR
        NEW.public_key_cose IS DISTINCT FROM OLD.public_key_cose OR
        NEW.aaguid IS DISTINCT FROM OLD.aaguid OR
        NEW.attestation_format IS DISTINCT FROM OLD.attestation_format OR
        NEW.attestation_type IS DISTINCT FROM OLD.attestation_type OR
        NEW.rp_id IS DISTINCT FROM OLD.rp_id OR
        NEW.allowed_origins IS DISTINCT FROM OLD.allowed_origins OR
        NEW.allowed_origins_sha256 IS DISTINCT FROM OLD.allowed_origins_sha256 OR
        NEW.user_verification IS DISTINCT FROM OLD.user_verification OR
        NEW.backup_eligible IS DISTINCT FROM OLD.backup_eligible OR
        NEW.backup_state IS DISTINCT FROM OLD.backup_state OR
        NEW.sign_count IS DISTINCT FROM OLD.sign_count OR
        NEW.counter_policy IS DISTINCT FROM OLD.counter_policy OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR NEW.revision<>OLD.revision+1 OR
        NOT ((OLD.state='pending' AND NEW.state='active') OR
             (OLD.state='active' AND NEW.state IN ('stale','deleted')))
    ) THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.iam_recovery_anchor_operation_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='recovery anchor operation deletion is forbidden';
    ELSIF TG_OP='INSERT' AND NOT (
        (NEW.kind='anchor_register' AND NEW.state='options_issued') OR
        (NEW.kind='anchor_delete' AND NEW.state='authority_pending')) THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='recovery anchor operation origin is invalid';
    ELSIF TG_OP='UPDATE' AND (
        NEW.operation_id IS DISTINCT FROM OLD.operation_id OR
        NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.tenant_id IS DISTINCT FROM OLD.tenant_id OR
        NEW.actor_user_id IS DISTINCT FROM OLD.actor_user_id OR
        NEW.kind IS DISTINCT FROM OLD.kind OR NEW.anchor_id IS DISTINCT FROM OLD.anchor_id OR
        NEW.options_idempotency_key IS DISTINCT FROM OLD.options_idempotency_key OR
        NEW.options_request_sha256 IS DISTINCT FROM OLD.options_request_sha256 OR
        NEW.challenge IS DISTINCT FROM OLD.challenge OR
        NEW.challenge_expires_at IS DISTINCT FROM OLD.challenge_expires_at OR
        NEW.rp_id IS DISTINCT FROM OLD.rp_id OR NEW.allowed_origins IS DISTINCT FROM OLD.allowed_origins OR
        NEW.allowed_origins_sha256 IS DISTINCT FROM OLD.allowed_origins_sha256 OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR NOT (
            (OLD.state='options_issued' AND NEW.state='authority_pending') OR
            (OLD.state='authority_pending' AND NEW.state='proposal_signed') OR
            (OLD.state='proposal_signed' AND NEW.state='completed'))
    ) THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor operation transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER iam_recovery_anchors_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_recovery_anchors
FOR EACH ROW EXECUTE FUNCTION platform.iam_recovery_anchor_guard_v1();
ALTER TABLE platform.iam_recovery_anchors ENABLE ALWAYS TRIGGER iam_recovery_anchors_guard;
CREATE TRIGGER iam_recovery_anchor_operations_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_recovery_anchor_operations
FOR EACH ROW EXECUTE FUNCTION platform.iam_recovery_anchor_operation_guard_v1();
ALTER TABLE platform.iam_recovery_anchor_operations
ENABLE ALWAYS TRIGGER iam_recovery_anchor_operations_guard;

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_operation_guard_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_list_v1(uuid,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_options_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,text,jsonb,bytea,timestamptz,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_registration_v1(uuid,uuid,uuid)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_authority_context_v1(uuid,uuid,uuid)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_stage_registration_v1(
    uuid,uuid,uuid,uuid,text,bytea,bytea,bytea,uuid,text,text,text,jsonb,bytea,bigint,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_stage_delete_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,timestamptz) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_record_proposal_v1(
    uuid,uuid,uuid,uuid,bigint,bytea,bytea,timestamptz) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_commit_v1(uuid,uuid,uuid,timestamptz)
TO harpy_runtime;
