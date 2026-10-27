-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_license_refresh_operation_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license refresh operation cannot be deleted';
    END IF;
    IF TG_OP = 'INSERT' THEN RETURN NEW; END IF;
    IF NEW.operation_id IS DISTINCT FROM OLD.operation_id
       OR NEW.license_id IS DISTINCT FROM OLD.license_id
       OR NEW.licensee_organization_id IS DISTINCT FROM OLD.licensee_organization_id
       OR NEW.environment_id IS DISTINCT FROM OLD.environment_id
       OR NEW.delivery_kind IS DISTINCT FROM OLD.delivery_kind
       OR NEW.nonce IS DISTINCT FROM OLD.nonce
       OR NEW.component_versions IS DISTINCT FROM OLD.component_versions
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.current_attempt < OLD.current_attempt
       OR NEW.current_attempt > OLD.current_attempt + 1 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license refresh operation identity is immutable';
    END IF;
    IF NOT (
        (OLD.state = 'pending' AND NEW.state = 'running')
        OR (OLD.state = 'running' AND NEW.state IN ('pending', 'succeeded', 'fenced'))
        OR (OLD.state = 'succeeded' AND NEW.state = 'running')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license refresh transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.billing_guard_license_refresh_attempt_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license refresh attempt cannot be deleted';
    END IF;
    IF TG_OP = 'INSERT' THEN RETURN NEW; END IF;
    IF OLD.state <> 'running' OR NEW.state = 'running'
       OR NEW.operation_id IS DISTINCT FROM OLD.operation_id
       OR NEW.attempt IS DISTINCT FROM OLD.attempt
       OR NEW.lease_token IS DISTINCT FROM OLD.lease_token
       OR NEW.claimed_at IS DISTINCT FROM OLD.claimed_at
       OR NEW.lease_until IS DISTINCT FROM OLD.lease_until THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license refresh attempt transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.billing_guard_license_head_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license head cannot be deleted';
    END IF;
    IF TG_OP = 'INSERT' THEN RETURN NEW; END IF;
    IF NEW.license_id IS DISTINCT FROM OLD.license_id
       OR NEW.environment_id IS DISTINCT FROM OLD.environment_id
       OR NEW.sequence <= OLD.sequence
       OR NEW.revision <> OLD.revision + 1
       OR NEW.snapshot_id IS NOT DISTINCT FROM OLD.snapshot_id
       OR NEW.payload_sha256 IS NOT DISTINCT FROM OLD.payload_sha256
       OR NEW.accepted_at < OLD.accepted_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license head must advance monotonically';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_guard_license_refresh_operation_v1()
FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.billing_guard_license_refresh_attempt_v1()
FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.billing_guard_license_head_v1()
FROM PUBLIC;

CREATE TRIGGER billing_license_refresh_operation_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_license_refresh_operations
FOR EACH ROW EXECUTE FUNCTION
    platform.billing_guard_license_refresh_operation_v1();
ALTER TABLE platform.billing_license_refresh_operations
    ENABLE ALWAYS TRIGGER billing_license_refresh_operation_guard;

CREATE TRIGGER billing_license_refresh_attempt_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_license_refresh_attempts
FOR EACH ROW EXECUTE FUNCTION
    platform.billing_guard_license_refresh_attempt_v1();
ALTER TABLE platform.billing_license_refresh_attempts
    ENABLE ALWAYS TRIGGER billing_license_refresh_attempt_guard;

CREATE TRIGGER billing_license_head_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_license_heads
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_license_head_v1();
ALTER TABLE platform.billing_license_heads
    ENABLE ALWAYS TRIGGER billing_license_head_guard;

CREATE TRIGGER billing_license_snapshot_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_license_snapshots
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_license_snapshots
    ENABLE ALWAYS TRIGGER billing_license_snapshot_immutable_guard;

CREATE TRIGGER billing_license_projection_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_license_snapshot_projections
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_license_snapshot_projections
    ENABLE ALWAYS TRIGGER billing_license_projection_immutable_guard;

CREATE TRIGGER billing_license_verification_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_license_verification_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_license_verification_receipts
    ENABLE ALWAYS TRIGGER billing_license_verification_immutable_guard;

CREATE TRIGGER billing_license_offering_publication_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_license_offering_publications
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_license_offering_publications
    ENABLE ALWAYS TRIGGER billing_license_offering_publication_immutable_guard;

CREATE TRIGGER billing_license_restore_receipt_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_license_restore_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_license_restore_receipts
    ENABLE ALWAYS TRIGGER billing_license_restore_receipt_immutable_guard;
