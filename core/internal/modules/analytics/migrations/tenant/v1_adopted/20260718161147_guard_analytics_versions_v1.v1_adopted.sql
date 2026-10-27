-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_guard_query_version_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics query version is immutable';
    END IF;
    IF NEW.id IS DISTINCT FROM OLD.id OR
       NEW.query_id IS DISTINCT FROM OLD.query_id OR
       NEW.version_no IS DISTINCT FROM OLD.version_no OR
       NEW.created_by IS DISTINCT FROM OLD.created_by OR
       NEW.created_at IS DISTINCT FROM OLD.created_at OR
       NEW.row_revision IS DISTINCT FROM OLD.row_revision + 1 OR
       NEW.updated_at < OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics query version identity is immutable';
    END IF;
    IF OLD.version_state = 'draft' THEN
        IF NEW.version_state NOT IN ('draft', 'published') THEN
            RAISE EXCEPTION USING ERRCODE = 'P6102',
                MESSAGE = 'analytics query draft transition is invalid';
        END IF;
        RETURN NEW;
    END IF;
    IF NOT ((OLD.version_state = 'published' AND
             NEW.version_state IN ('needs_revalidation', 'retired')) OR
            (OLD.version_state = 'needs_revalidation' AND
             NEW.version_state = 'retired')) OR
       NEW.name IS DISTINCT FROM OLD.name OR
       NEW.query_document IS DISTINCT FROM OLD.query_document OR
       NEW.query_sha256 IS DISTINCT FROM OLD.query_sha256 OR
       NEW.parameter_document IS DISTINCT FROM OLD.parameter_document OR
       NEW.parameter_sha256 IS DISTINCT FROM OLD.parameter_sha256 OR
       NEW.output_document IS DISTINCT FROM OLD.output_document OR
       NEW.output_sha256 IS DISTINCT FROM OLD.output_sha256 OR
       NEW.manifest_sha256 IS DISTINCT FROM OLD.manifest_sha256 OR
       NEW.cache_policy IS DISTINCT FROM OLD.cache_policy OR
       NEW.published_by IS DISTINCT FROM OLD.published_by OR
       NEW.published_at IS DISTINCT FROM OLD.published_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics published query version is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_guard_dashboard_version_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics dashboard version is immutable';
    END IF;
    IF NEW.id IS DISTINCT FROM OLD.id OR
       NEW.dashboard_id IS DISTINCT FROM OLD.dashboard_id OR
       NEW.version_no IS DISTINCT FROM OLD.version_no OR
       NEW.created_by IS DISTINCT FROM OLD.created_by OR
       NEW.created_at IS DISTINCT FROM OLD.created_at OR
       NEW.row_revision IS DISTINCT FROM OLD.row_revision + 1 OR
       NEW.updated_at < OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics dashboard version identity is immutable';
    END IF;
    IF OLD.version_state = 'draft' THEN
        IF NEW.version_state NOT IN ('draft', 'published') THEN
            RAISE EXCEPTION USING ERRCODE = 'P6102',
                MESSAGE = 'analytics dashboard draft transition is invalid';
        END IF;
        RETURN NEW;
    END IF;
    IF NOT (OLD.version_state = 'published' AND NEW.version_state = 'retired') OR
       NEW.name IS DISTINCT FROM OLD.name OR
       NEW.visibility IS DISTINCT FROM OLD.visibility OR
       NEW.layout_document IS DISTINCT FROM OLD.layout_document OR
       NEW.layout_sha256 IS DISTINCT FROM OLD.layout_sha256 OR
       NEW.manifest_sha256 IS DISTINCT FROM OLD.manifest_sha256 OR
       NEW.published_by IS DISTINCT FROM OLD.published_by OR
       NEW.published_at IS DISTINCT FROM OLD.published_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics published dashboard version is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_guard_dashboard_widget_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_version_id uuid;
    v_state text;
BEGIN
    v_version_id := CASE WHEN TG_OP = 'INSERT'
        THEN NEW.dashboard_version_id ELSE OLD.dashboard_version_id END;
    SELECT version_row.version_state INTO STRICT v_state
    FROM analytics_dashboard_versions AS version_row
    WHERE version_row.id = v_version_id;
    IF v_state IS DISTINCT FROM 'draft' OR
       (TG_OP = 'UPDATE' AND NEW.dashboard_version_id IS DISTINCT FROM v_version_id) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics published dashboard widget is immutable';
    END IF;
    RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_guard_query_version_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_dashboard_version_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_dashboard_widget_v1() FROM PUBLIC;

CREATE TRIGGER analytics_saved_query_versions_guard
BEFORE UPDATE OR DELETE ON analytics_saved_query_versions
FOR EACH ROW EXECUTE FUNCTION analytics_guard_query_version_v1();
CREATE TRIGGER analytics_dashboard_versions_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_versions
FOR EACH ROW EXECUTE FUNCTION analytics_guard_dashboard_version_v1();
CREATE TRIGGER analytics_dashboard_widgets_guard
BEFORE INSERT OR UPDATE OR DELETE ON analytics_dashboard_widgets
FOR EACH ROW EXECUTE FUNCTION analytics_guard_dashboard_widget_v1();

ALTER TABLE analytics_saved_query_versions
    ENABLE ALWAYS TRIGGER analytics_saved_query_versions_guard;
ALTER TABLE analytics_dashboard_versions
    ENABLE ALWAYS TRIGGER analytics_dashboard_versions_guard;
ALTER TABLE analytics_dashboard_widgets
    ENABLE ALWAYS TRIGGER analytics_dashboard_widgets_guard;
