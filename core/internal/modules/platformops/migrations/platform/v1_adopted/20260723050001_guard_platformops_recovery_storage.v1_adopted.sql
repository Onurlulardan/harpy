-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_reject_append_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION 'platform operations evidence is append-only'
        USING ERRCODE = '55000';
END;
$function$;

CREATE TRIGGER platformops_recovery_signing_append_only
BEFORE UPDATE OR DELETE ON platform.platformops_recovery_signing_bundles
FOR EACH ROW EXECUTE FUNCTION platform.platformops_reject_append_mutation_v1();

CREATE TRIGGER platformops_recovery_components_append_only
BEFORE UPDATE OR DELETE ON platform.platformops_recovery_components
FOR EACH ROW EXECUTE FUNCTION platform.platformops_reject_append_mutation_v1();

CREATE FUNCTION platform.platformops_guard_recovery_set_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NEW.recovery_set_id <> OLD.recovery_set_id
       OR NEW.operation_id <> OLD.operation_id
       OR NEW.environment_id <> OLD.environment_id
       OR NEW.deployment_mode <> OLD.deployment_mode
       OR NEW.portability <> OLD.portability
       OR NEW.sequence <> OLD.sequence
       OR NEW.primary_binding_id <> OLD.primary_binding_id
       OR NEW.tenant_set_sha256 <> OLD.tenant_set_sha256
       OR NEW.owner_binding_id <> OLD.owner_binding_id
       OR NEW.owner_binding_generation <> OLD.owner_binding_generation
       OR NEW.database_lineage_id <> OLD.database_lineage_id
       OR NEW.recovery_signing_generation <> OLD.recovery_signing_generation
       OR NEW.recovery_trust_bundle_sha256 <> OLD.recovery_trust_bundle_sha256
       OR NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'recovery set identity is immutable'
            USING ERRCODE = '55000';
    END IF;
    IF OLD.state <> NEW.state AND NOT (
        (OLD.state = 'requested' AND NEW.state IN ('holding','failed'))
        OR (OLD.state = 'holding' AND NEW.state IN ('snapshotting','failed'))
        OR (OLD.state = 'snapshotting' AND NEW.state IN ('copying','failed'))
        OR (OLD.state = 'copying' AND NEW.state IN ('verifying','failed'))
        OR (OLD.state = 'verifying' AND NEW.state IN ('protected','failed'))
        OR (OLD.state = 'protected' AND NEW.state = 'deleting')
        OR (OLD.state = 'deleting' AND NEW.state IN ('protected','deleted'))
    ) THEN
        RAISE EXCEPTION 'recovery set transition is invalid'
            USING ERRCODE = '55000';
    END IF;
    IF OLD.state = 'protected' AND NEW.state <> 'deleting' AND (
        NEW.manifest_sha256 IS DISTINCT FROM OLD.manifest_sha256
        OR NEW.manifest_bytes IS DISTINCT FROM OLD.manifest_bytes
        OR NEW.database_system_id IS DISTINCT FROM OLD.database_system_id
        OR NEW.protected_at IS DISTINCT FROM OLD.protected_at
    ) THEN
        RAISE EXCEPTION 'protected recovery evidence is immutable'
            USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE TRIGGER platformops_guard_recovery_set
BEFORE UPDATE ON platform.platformops_recovery_sets
FOR EACH ROW EXECUTE FUNCTION platform.platformops_guard_recovery_set_v1();

CREATE FUNCTION platform.platformops_guard_restore_operation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NEW.operation_id <> OLD.operation_id
       OR NEW.recovery_set_id <> OLD.recovery_set_id
       OR NEW.environment_id <> OLD.environment_id
       OR NEW.target_disposition <> OLD.target_disposition
       OR NEW.plan_sha256 <> OLD.plan_sha256
       OR NEW.plan_bytes <> OLD.plan_bytes
       OR NEW.safety_recovery_set_id IS DISTINCT FROM OLD.safety_recovery_set_id
       OR NEW.safety_recovery_set_sha256 IS DISTINCT FROM OLD.safety_recovery_set_sha256
       OR NEW.waiver_reason IS DISTINCT FROM OLD.waiver_reason
       OR NEW.probe_sha256 IS DISTINCT FROM OLD.probe_sha256
       OR NEW.destructive_confirmed <> OLD.destructive_confirmed
       OR NEW.seeded OR NEW.owner_changed OR NEW.tenant_replaced
       OR NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'restore operation identity is immutable'
            USING ERRCODE = '55000';
    END IF;
    IF OLD.state <> NEW.state AND NOT (
        (OLD.state = 'plan_ready' AND NEW.state IN ('draining','failed','fenced'))
        OR (OLD.state = 'draining' AND NEW.state IN ('restoring','failed','fenced'))
        OR (OLD.state = 'restoring' AND NEW.state IN ('verifying','failed','fenced'))
        OR (OLD.state = 'verifying' AND NEW.state IN ('switched','failed','fenced'))
        OR (OLD.state = 'switched' AND NEW.state = 'completed')
    ) THEN
        RAISE EXCEPTION 'restore operation transition is invalid'
            USING ERRCODE = '55000';
    END IF;
    IF OLD.state = 'completed' THEN
        RAISE EXCEPTION 'completed restore is immutable'
            USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER platformops_guard_restore_operation
BEFORE UPDATE ON platform.platformops_restore_operations
FOR EACH ROW EXECUTE FUNCTION platform.platformops_guard_restore_operation_v1();

CREATE TRIGGER platformops_recovery_sets_no_delete
BEFORE DELETE ON platform.platformops_recovery_sets
FOR EACH ROW EXECUTE FUNCTION platform.platformops_reject_append_mutation_v1();

CREATE TRIGGER platformops_restore_operations_no_delete
BEFORE DELETE ON platform.platformops_restore_operations
FOR EACH ROW EXECUTE FUNCTION platform.platformops_reject_append_mutation_v1();
