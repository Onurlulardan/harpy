-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_guard_onboarding_operation_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'onboarding operation delete is forbidden';
    END IF;
    IF ROW(NEW.operation_id, NEW.environment_id, NEW.operation_kind,
        NEW.idempotency_key, NEW.request_sha256, NEW.manifest_sha256,
        NEW.step_plan_sha256, NEW.deployment_mode, NEW.tenant_id,
        NEW.schema_name, NEW.binding_purpose, NEW.first_user_id,
        NEW.root_organization_id, NEW.protected_admin_binding_id,
        NEW.instance_owner_binding_id, NEW.offering_version_id,
        NEW.offering_revision, NEW.offering_sha256, NEW.offering_effective_at,
        NEW.offering_source_sequence, NEW.offering_source_sha256,
        NEW.recovery_signing_generation, NEW.recovery_trust_bundle_sha256,
        NEW.registry_binding_id, NEW.authority_scope_kind, NEW.authority_scope_id,
        NEW.authority_generation, NEW.authority_previous_head_sha256,
        NEW.authority_state_sha256, NEW.authority_proposal_sha256, NEW.created_at)
       IS DISTINCT FROM
       ROW(OLD.operation_id, OLD.environment_id, OLD.operation_kind,
        OLD.idempotency_key, OLD.request_sha256, OLD.manifest_sha256,
        OLD.step_plan_sha256, OLD.deployment_mode, OLD.tenant_id,
        OLD.schema_name, OLD.binding_purpose, OLD.first_user_id,
        OLD.root_organization_id, OLD.protected_admin_binding_id,
        OLD.instance_owner_binding_id, OLD.offering_version_id,
        OLD.offering_revision, OLD.offering_sha256, OLD.offering_effective_at,
        OLD.offering_source_sequence, OLD.offering_source_sha256,
        OLD.recovery_signing_generation, OLD.recovery_trust_bundle_sha256,
        OLD.registry_binding_id, OLD.authority_scope_kind, OLD.authority_scope_id,
        OLD.authority_generation, OLD.authority_previous_head_sha256,
        OLD.authority_state_sha256, OLD.authority_proposal_sha256, OLD.created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'onboarding plan is immutable';
    END IF;
    IF NEW.revision <> OLD.revision + 1 OR NEW.updated_at <= OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'onboarding revision transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.tenancy_reject_onboarding_append_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'onboarding evidence is append only';
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER tenancy_onboarding_operation_guard
BEFORE UPDATE OR DELETE ON platform.tenancy_onboarding_operations
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_guard_onboarding_operation_v1();

CREATE TRIGGER tenancy_onboarding_steps_guard
BEFORE UPDATE OR DELETE ON platform.tenancy_onboarding_steps
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_onboarding_append_mutation_v1();

CREATE TRIGGER tenancy_onboarding_receipts_guard
BEFORE UPDATE OR DELETE ON platform.tenancy_onboarding_receipts
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_onboarding_append_mutation_v1();

CREATE TRIGGER tenancy_onboarding_authority_acks_guard
BEFORE UPDATE OR DELETE ON platform.tenancy_onboarding_authority_acks
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_onboarding_append_mutation_v1();
