-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_candidate_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM analytics_catalog_version_projections AS projection
        JOIN analytics_catalog_projection_attempts AS attempt
          ON attempt.receipt_id=NEW.attempt_receipt_id
         AND attempt.projection_receipt_id=projection.receipt_id
        WHERE projection.receipt_id=NEW.projection_receipt_id
          AND projection.component_id=NEW.component_id
          AND projection.package_id=NEW.package_id
          AND projection.package_version=NEW.package_version
          AND projection.release_id=NEW.release_id
          AND projection.owner_generation=NEW.owner_generation
          AND projection.content_sha256=NEW.content_sha256
          AND projection.operational_state='ready'
          AND projection.ready_at IS NOT NULL
          AND attempt.ready_at IS NOT NULL
          AND projection.ready_at<=NEW.prepared_at
          AND attempt.updated_at<=NEW.prepared_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog candidate lacks exact ready evidence';
    END IF;
    IF TG_OP='UPDATE' AND (
        ROW(
            NEW.component_id,NEW.projection_receipt_id,NEW.package_id,
            NEW.package_version,NEW.release_id,NEW.owner_generation,
            NEW.content_sha256
        ) IS DISTINCT FROM ROW(
            OLD.component_id,OLD.projection_receipt_id,OLD.package_id,
            OLD.package_version,OLD.release_id,OLD.owner_generation,
            OLD.content_sha256
        ) OR NEW.prepared_at<OLD.prepared_at OR
        (NEW.attempt_receipt_id=OLD.attempt_receipt_id AND
         OLD.staged_operation_id IS NOT NULL AND
         NEW.staged_operation_id IS DISTINCT FROM OLD.staged_operation_id) OR
        (NEW.attempt_receipt_id<>OLD.attempt_receipt_id AND
         NEW.staged_operation_id IS NOT NULL)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog candidate update differs';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_catalog_activation_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP='DELETE' OR ROW(
        NEW.projection_receipt_id,NEW.attempt_receipt_id,
        NEW.component_id,NEW.package_id,NEW.package_version,
        NEW.release_id,NEW.content_sha256,NEW.activated_at
    ) IS DISTINCT FROM ROW(
        OLD.projection_receipt_id,OLD.attempt_receipt_id,
        OLD.component_id,OLD.package_id,OLD.package_version,
        OLD.release_id,OLD.content_sha256,OLD.activated_at
    ) OR
       (OLD.superseded_at IS NOT NULL AND NEW IS DISTINCT FROM OLD) OR
       (OLD.superseded_at IS NULL AND NEW.superseded_at IS NOT NULL AND
        NEW.superseded_at<OLD.activated_at) OR
       (OLD.superseded_at IS NULL AND NEW.superseded_at IS NULL AND
        NEW IS DISTINCT FROM OLD) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog activation evidence is immutable';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_candidate_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_activation_guard_v1() FROM PUBLIC;

CREATE TRIGGER analytics_catalog_component_candidates_ready_guard
BEFORE INSERT OR UPDATE ON analytics_catalog_component_candidates
FOR EACH ROW EXECUTE FUNCTION analytics_catalog_candidate_guard_v1();
CREATE TRIGGER analytics_catalog_projection_activations_immutable
BEFORE UPDATE OR DELETE ON analytics_catalog_projection_activations
FOR EACH ROW EXECUTE FUNCTION analytics_catalog_activation_guard_v1();

ALTER TABLE analytics_catalog_component_candidates
    ENABLE ALWAYS TRIGGER analytics_catalog_component_candidates_ready_guard;
ALTER TABLE analytics_catalog_projection_activations
    ENABLE ALWAYS TRIGGER analytics_catalog_projection_activations_immutable;
