-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_demo_ready_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NEW.ready AND NOT OLD.ready AND NOT EXISTS (
        SELECT 1
        FROM catalog_demo_handoffs AS handoff
        WHERE handoff.provision_receipt_id = NEW.receipt_id
          AND handoff.component_id = NEW.component_id
          AND handoff.package_id = NEW.package_id
          AND handoff.package_version = NEW.package_version
          AND handoff.release_id = NEW.release_id
          AND handoff.content_sha256 = NEW.content_sha256
          AND handoff.request_sha256 = NEW.request_sha256
          AND handoff.completed_at = NEW.completed_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog demo ready transition lacks Forms handoff';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION catalog_demo_activation_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NEW.component_kind = 'DEMO_DATA' AND
       (NEW.lifecycle_state IS DISTINCT FROM OLD.lifecycle_state OR
        NEW.content_sha256 IS DISTINCT FROM OLD.content_sha256) AND
       NEW.lifecycle_state = 'STAGED' AND NOT EXISTS (
        SELECT 1
        FROM catalog_demo_provisions AS provision
        JOIN catalog_demo_handoffs AS handoff
          ON handoff.provision_receipt_id = provision.receipt_id
        WHERE provision.component_id = NEW.component_id
          AND provision.owner_generation = NEW.owner_generation
          AND provision.contract = NEW.contract
          AND provision.content_sha256 = NEW.content_sha256
          AND provision.ready
          AND provision.retry_at IS NULL
          AND handoff.completed_at = provision.completed_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog demo activation lacks Forms handoff readback';
    END IF;
    IF NEW.component_kind = 'DEMO_DATA' AND
       (NEW.lifecycle_state IS DISTINCT FROM OLD.lifecycle_state OR
        NEW.content_sha256 IS DISTINCT FROM OLD.content_sha256) AND
       NEW.lifecycle_state IN ('ACTIVE','DISABLED') AND NOT EXISTS (
        SELECT 1
        FROM catalog_demo_component_heads AS head
        JOIN catalog_demo_provisions AS provision
          ON provision.receipt_id = head.provision_receipt_id
         AND provision.component_id = head.component_id
         AND provision.package_id = head.package_id
         AND provision.package_version = head.package_version
         AND provision.release_id = head.release_id
         AND provision.content_sha256 = head.content_sha256
        JOIN catalog_demo_handoffs AS handoff
          ON handoff.handoff_id = head.handoff_id
         AND handoff.provision_receipt_id = provision.receipt_id
         AND handoff.forms_run_id = head.forms_run_id
        WHERE head.component_id = NEW.component_id
          AND head.content_sha256 = NEW.content_sha256
          AND provision.ready AND provision.retry_at IS NULL
          AND handoff.completed_at = provision.completed_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog demo active head lacks exact ready handoff';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION catalog_demo_head_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM catalog_demo_provisions AS provision
        JOIN catalog_demo_handoffs AS handoff
          ON handoff.handoff_id = NEW.handoff_id
         AND handoff.provision_receipt_id = provision.receipt_id
         AND handoff.component_id = provision.component_id
         AND handoff.package_id = provision.package_id
         AND handoff.package_version = provision.package_version
         AND handoff.release_id = provision.release_id
         AND handoff.forms_run_id = NEW.forms_run_id
         AND handoff.content_sha256 = provision.content_sha256
        WHERE provision.receipt_id = NEW.provision_receipt_id
          AND provision.component_id = NEW.component_id
          AND provision.package_id = NEW.package_id
          AND provision.package_version = NEW.package_version
          AND provision.release_id = NEW.release_id
          AND provision.content_sha256 = NEW.content_sha256
          AND provision.ready AND provision.retry_at IS NULL
          AND handoff.completed_at = provision.completed_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog demo head lacks exact ready handoff';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION catalog_demo_evidence_immutable_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Catalog demo evidence is immutable';
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION catalog_demo_ready_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION catalog_demo_activation_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION catalog_demo_head_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION catalog_demo_evidence_immutable_v1() FROM PUBLIC;

CREATE TRIGGER catalog_demo_provisions_ready_guard
BEFORE UPDATE OF ready ON catalog_demo_provisions
FOR EACH ROW EXECUTE FUNCTION catalog_demo_ready_guard_v1();
CREATE TRIGGER catalog_demo_components_activation_guard
BEFORE UPDATE OF lifecycle_state, content_sha256 ON catalog_catalog_components
FOR EACH ROW EXECUTE FUNCTION catalog_demo_activation_guard_v1();
CREATE TRIGGER catalog_demo_component_heads_ready_guard
BEFORE INSERT OR UPDATE ON catalog_demo_component_heads
FOR EACH ROW EXECUTE FUNCTION catalog_demo_head_guard_v1();
CREATE TRIGGER catalog_demo_payloads_immutable
BEFORE UPDATE OR DELETE ON catalog_demo_payloads
FOR EACH ROW EXECUTE FUNCTION catalog_demo_evidence_immutable_v1();
CREATE TRIGGER catalog_demo_handoffs_immutable
BEFORE UPDATE OR DELETE ON catalog_demo_handoffs
FOR EACH ROW EXECUTE FUNCTION catalog_demo_evidence_immutable_v1();

ALTER TABLE catalog_demo_provisions
    ENABLE ALWAYS TRIGGER catalog_demo_provisions_ready_guard;
ALTER TABLE catalog_catalog_components
    ENABLE ALWAYS TRIGGER catalog_demo_components_activation_guard;
ALTER TABLE catalog_demo_component_heads
    ENABLE ALWAYS TRIGGER catalog_demo_component_heads_ready_guard;
ALTER TABLE catalog_demo_payloads
    ENABLE ALWAYS TRIGGER catalog_demo_payloads_immutable;
ALTER TABLE catalog_demo_handoffs
    ENABLE ALWAYS TRIGGER catalog_demo_handoffs_immutable;
