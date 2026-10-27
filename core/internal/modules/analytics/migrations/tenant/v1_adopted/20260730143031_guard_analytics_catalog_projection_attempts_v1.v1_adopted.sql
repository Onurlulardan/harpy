-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_evidence_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog projection evidence is immutable';
    END IF;
    IF ROW(
        NEW.receipt_id,NEW.package_id,NEW.package_slug,NEW.package_version,
        NEW.release_id,NEW.component_id,NEW.owner_generation,NEW.contract,
        NEW.content_sha256,NEW.component_kind,NEW.analytics_kind,
        NEW.module_slug,NEW.semantic_identity,NEW.semantic_sha256,
        NEW.item_count,NEW.auxiliary_count,NEW.name,NEW.description,
        NEW.legacy_hash,NEW.definition_json,NEW.definition_sha256,
        NEW.document_bytes,NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.receipt_id,OLD.package_id,OLD.package_slug,OLD.package_version,
        OLD.release_id,OLD.component_id,OLD.owner_generation,OLD.contract,
        OLD.content_sha256,OLD.component_kind,OLD.analytics_kind,
        OLD.module_slug,OLD.semantic_identity,OLD.semantic_sha256,
        OLD.item_count,OLD.auxiliary_count,OLD.name,OLD.description,
        OLD.legacy_hash,OLD.definition_json,OLD.definition_sha256,
        OLD.document_bytes,OLD.created_at
    ) OR NEW.updated_at<OLD.updated_at OR
       (OLD.operational_state='ready' AND NEW IS DISTINCT FROM OLD) OR
       (OLD.operational_state='pending' AND
        NEW.operational_state NOT IN ('pending','ready')) OR
       (OLD.operational_state='pending' AND
        NEW.operational_state='pending' AND ROW(
            NEW.operational_resource_id,NEW.published_version_id,
            NEW.current_draft_version_id,NEW.ready_at
        ) IS DISTINCT FROM ROW(
            OLD.operational_resource_id,OLD.published_version_id,
            OLD.current_draft_version_id,OLD.ready_at
        )) OR
       (OLD.operational_state='pending' AND
        NEW.operational_state='ready' AND (
            ROW(NEW.runtime_plan_bytes,NEW.runtime_plan_sha256)
              IS DISTINCT FROM
            ROW(OLD.runtime_plan_bytes,OLD.runtime_plan_sha256) OR
            NEW.ready_at<OLD.updated_at OR NEW.updated_at<>NEW.ready_at
        )) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog projection evidence update differs';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_catalog_attempt_evidence_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog projection attempt is immutable';
    END IF;
    IF ROW(
        NEW.receipt_id,NEW.projection_receipt_id,NEW.actor_id,
        NEW.idempotency_key,NEW.request_sha256,NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.receipt_id,OLD.projection_receipt_id,OLD.actor_id,
        OLD.idempotency_key,OLD.request_sha256,OLD.created_at
    ) OR NEW.updated_at<OLD.updated_at OR
       (OLD.ready_at IS NOT NULL AND NEW IS DISTINCT FROM OLD) OR
       (OLD.ready_at IS NULL AND NEW.ready_at IS NOT NULL AND (
            NEW.ready_at<OLD.updated_at OR NEW.updated_at<>NEW.ready_at
       )) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog projection attempt update differs';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_projection_evidence_guard_v1()
    FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_attempt_evidence_guard_v1()
    FROM PUBLIC;

CREATE TRIGGER analytics_catalog_version_projections_immutable
BEFORE UPDATE OR DELETE ON analytics_catalog_version_projections
FOR EACH ROW EXECUTE FUNCTION analytics_catalog_projection_evidence_guard_v1();
CREATE TRIGGER analytics_catalog_projection_attempts_immutable
BEFORE UPDATE OR DELETE ON analytics_catalog_projection_attempts
FOR EACH ROW EXECUTE FUNCTION analytics_catalog_attempt_evidence_guard_v1();

ALTER TABLE analytics_catalog_version_projections
    ENABLE ALWAYS TRIGGER analytics_catalog_version_projections_immutable;
ALTER TABLE analytics_catalog_projection_attempts
    ENABLE ALWAYS TRIGGER analytics_catalog_projection_attempts_immutable;
