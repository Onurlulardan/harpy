-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_head_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM analytics_catalog_version_projections AS projection
        JOIN analytics_catalog_projection_activations AS activation
          ON activation.projection_receipt_id=projection.receipt_id
         AND activation.attempt_receipt_id IS NOT NULL
         AND activation.component_id=projection.component_id
         AND activation.package_id=projection.package_id
         AND activation.package_version=projection.package_version
         AND activation.release_id=projection.release_id
         AND activation.content_sha256=projection.content_sha256
         AND activation.superseded_at IS NULL
        JOIN analytics_catalog_projection_attempts AS attempt
          ON attempt.receipt_id=activation.attempt_receipt_id
         AND attempt.projection_receipt_id=projection.receipt_id
         AND attempt.ready_at IS NOT NULL
        JOIN analytics_catalog_definitions AS definition
          ON definition.component_id=projection.component_id
         AND definition.package_id=projection.package_id
         AND definition.package_version=projection.package_version
         AND definition.release_id=projection.release_id
         AND definition.content_sha256=projection.content_sha256
         AND definition.operational_resource_id=
             projection.operational_resource_id
         AND definition.published_version_id=projection.published_version_id
         AND definition.current_draft_version_id=
             projection.current_draft_version_id
        WHERE projection.receipt_id=NEW.projection_receipt_id
          AND projection.component_id=NEW.component_id
          AND projection.package_id=NEW.package_id
          AND projection.package_version=NEW.package_version
          AND projection.release_id=NEW.release_id
          AND projection.content_sha256=NEW.content_sha256
          AND projection.operational_state='ready'
          AND projection.ready_at<=NEW.activated_at
          AND activation.activated_at=NEW.activated_at
          AND attempt.updated_at<=NEW.activated_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog head lacks exact live activation';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_catalog_component_head_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF (NEW.lifecycle_state IS DISTINCT FROM OLD.lifecycle_state OR
        NEW.content_sha256 IS DISTINCT FROM OLD.content_sha256) AND
       NEW.lifecycle_state IN ('ACTIVE','DISABLED','RETIRED') AND
       NOT EXISTS (
        SELECT 1
        FROM analytics_catalog_component_heads AS head
        JOIN analytics_catalog_version_projections AS projection
          ON projection.receipt_id=head.projection_receipt_id
        WHERE head.component_id=NEW.component_id
          AND head.content_sha256=NEW.content_sha256
          AND (
            (projection.analytics_kind='endpoint' AND EXISTS (
                SELECT 1 FROM analytics_saved_queries AS root
                WHERE root.id=projection.operational_resource_id
                  AND root.lifecycle_state=CASE
                    WHEN NEW.lifecycle_state='ACTIVE' THEN 'active'
                    ELSE 'retired' END
            )) OR
            (projection.analytics_kind='dashboard' AND EXISTS (
                SELECT 1 FROM analytics_dashboards AS root
                WHERE root.id=projection.operational_resource_id
                  AND root.lifecycle_state=CASE
                    WHEN NEW.lifecycle_state='ACTIVE' THEN 'active'
                    ELSE 'retired' END
            ))
          )
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog component lacks exact visible head';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_head_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_component_head_guard_v1()
    FROM PUBLIC;

CREATE TRIGGER analytics_catalog_component_heads_ready_guard
BEFORE INSERT OR UPDATE ON analytics_catalog_component_heads
FOR EACH ROW EXECUTE FUNCTION analytics_catalog_head_guard_v1();
CREATE TRIGGER analytics_catalog_components_head_guard
BEFORE UPDATE OF lifecycle_state,content_sha256
ON analytics_catalog_components
FOR EACH ROW EXECUTE FUNCTION analytics_catalog_component_head_guard_v1();

ALTER TABLE analytics_catalog_component_heads
    ENABLE ALWAYS TRIGGER analytics_catalog_component_heads_ready_guard;
ALTER TABLE analytics_catalog_components
    ENABLE ALWAYS TRIGGER analytics_catalog_components_head_guard;
