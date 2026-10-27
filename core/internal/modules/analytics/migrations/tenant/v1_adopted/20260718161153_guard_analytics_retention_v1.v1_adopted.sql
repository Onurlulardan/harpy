-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_retention_delete_allowed_v1(p_backend_pid integer)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_backend_pid IS DISTINCT FROM pg_catalog.pg_backend_pid() THEN
        RETURN false;
    END IF;
    RETURN EXISTS (
        SELECT 1 FROM analytics_retention_contexts AS retention_context
        WHERE retention_context.backend_pid = pg_catalog.pg_backend_pid()
          AND retention_context.transaction_id = pg_catalog.pg_current_xact_id()
    ) AND NOT EXISTS (
        SELECT 1 FROM analytics_legal_holds AS legal_hold
        WHERE legal_hold.state = 'active'
    );
END;
$function$;

CREATE FUNCTION analytics_reject_mutation_retention_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP = 'DELETE' AND analytics_retention_delete_allowed_v1(pg_catalog.pg_backend_pid()) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'analytics immutable row cannot be changed';
END;
$function$;

CREATE FUNCTION analytics_guard_operation_retention_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP = 'INSERT' THEN
        RETURN NEW;
    END IF;
    IF TG_OP = 'DELETE' AND analytics_retention_delete_allowed_v1(pg_catalog.pg_backend_pid()) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'analytics operation receipt is immutable';
END;
$function$;

CREATE FUNCTION analytics_guard_query_version_retention_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF analytics_retention_delete_allowed_v1(pg_catalog.pg_backend_pid()) THEN RETURN OLD; END IF;
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics query version is immutable';
    END IF;
    IF NEW.id IS DISTINCT FROM OLD.id OR NEW.query_id IS DISTINCT FROM OLD.query_id OR
       NEW.version_no IS DISTINCT FROM OLD.version_no OR
       NEW.created_by IS DISTINCT FROM OLD.created_by OR
       NEW.created_at IS DISTINCT FROM OLD.created_at OR
       NEW.row_revision IS DISTINCT FROM OLD.row_revision + 1 OR
       NEW.updated_at < OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics query version identity is immutable';
    END IF;
    IF OLD.version_state = 'draft' THEN
        IF NEW.version_state NOT IN ('draft','published') THEN
            RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics query draft transition is invalid';
        END IF;
        RETURN NEW;
    END IF;
    IF NOT ((OLD.version_state = 'published' AND
             NEW.version_state IN ('needs_revalidation','retired')) OR
            (OLD.version_state = 'needs_revalidation' AND NEW.version_state = 'retired')) OR
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
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics published query version is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_guard_dashboard_version_retention_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF analytics_retention_delete_allowed_v1(pg_catalog.pg_backend_pid()) THEN RETURN OLD; END IF;
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics dashboard version is immutable';
    END IF;
    IF NEW.id IS DISTINCT FROM OLD.id OR NEW.dashboard_id IS DISTINCT FROM OLD.dashboard_id OR
       NEW.version_no IS DISTINCT FROM OLD.version_no OR
       NEW.created_by IS DISTINCT FROM OLD.created_by OR
       NEW.created_at IS DISTINCT FROM OLD.created_at OR
       NEW.row_revision IS DISTINCT FROM OLD.row_revision + 1 OR
       NEW.updated_at < OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics dashboard version identity is immutable';
    END IF;
    IF OLD.version_state = 'draft' THEN
        IF NEW.version_state NOT IN ('draft','published') THEN
            RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics dashboard draft transition is invalid';
        END IF;
        RETURN NEW;
    END IF;
    IF NOT (OLD.version_state = 'published' AND NEW.version_state = 'retired') OR
       NEW.name IS DISTINCT FROM OLD.name OR NEW.visibility IS DISTINCT FROM OLD.visibility OR
       NEW.layout_document IS DISTINCT FROM OLD.layout_document OR
       NEW.layout_sha256 IS DISTINCT FROM OLD.layout_sha256 OR
       NEW.manifest_sha256 IS DISTINCT FROM OLD.manifest_sha256 OR
       NEW.published_by IS DISTINCT FROM OLD.published_by OR
       NEW.published_at IS DISTINCT FROM OLD.published_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics published dashboard version is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_guard_dashboard_widget_retention_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_version_id uuid; v_state text;
BEGIN
    IF TG_OP = 'DELETE' AND analytics_retention_delete_allowed_v1(pg_catalog.pg_backend_pid()) THEN RETURN OLD; END IF;
    v_version_id := CASE WHEN TG_OP = 'INSERT' THEN NEW.dashboard_version_id ELSE OLD.dashboard_version_id END;
    SELECT version_state INTO STRICT v_state FROM analytics_dashboard_versions WHERE id = v_version_id;
    IF v_state IS DISTINCT FROM 'draft' OR
       (TG_OP = 'UPDATE' AND NEW.dashboard_version_id IS DISTINCT FROM v_version_id) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics published dashboard widget is immutable';
    END IF;
    RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$function$;

CREATE FUNCTION analytics_guard_inbound_reference_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' OR OLD.state = 'released' OR NEW.reference_id IS DISTINCT FROM OLD.reference_id OR
       NEW.target_kind IS DISTINCT FROM OLD.target_kind OR NEW.target_id IS DISTINCT FROM OLD.target_id OR
       NEW.source_module IS DISTINCT FROM OLD.source_module OR NEW.source_kind IS DISTINCT FROM OLD.source_kind OR
       NEW.source_id IS DISTINCT FROM OLD.source_id OR NEW.source_version_id IS DISTINCT FROM OLD.source_version_id OR
       NEW.usage IS DISTINCT FROM OLD.usage OR NEW.created_at IS DISTINCT FROM OLD.created_at OR
       NEW.state IS DISTINCT FROM 'released' OR NEW.released_by IS NULL OR NEW.released_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'analytics inbound reference is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_guard_legal_hold_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' OR OLD.state = 'released' OR NEW.hold_id IS DISTINCT FROM OLD.hold_id OR
       NEW.reason_sha256 IS DISTINCT FROM OLD.reason_sha256 OR NEW.placed_by IS DISTINCT FROM OLD.placed_by OR
       NEW.placed_at IS DISTINCT FROM OLD.placed_at OR NEW.state IS DISTINCT FROM 'released' OR
       NEW.row_revision IS DISTINCT FROM OLD.row_revision + 1 OR NEW.released_by IS NULL OR NEW.released_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'analytics legal hold is immutable';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_retention_delete_allowed_v1(integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_reject_mutation_retention_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_operation_retention_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_query_version_retention_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_dashboard_version_retention_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_dashboard_widget_retention_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_inbound_reference_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_legal_hold_v1() FROM PUBLIC;
CREATE TRIGGER analytics_inbound_references_guard BEFORE UPDATE OR DELETE ON analytics_inbound_references
FOR EACH ROW EXECUTE FUNCTION analytics_guard_inbound_reference_v1();
CREATE TRIGGER analytics_legal_holds_guard BEFORE UPDATE OR DELETE ON analytics_legal_holds
FOR EACH ROW EXECUTE FUNCTION analytics_guard_legal_hold_v1();
CREATE TRIGGER analytics_retention_receipts_immutable BEFORE UPDATE OR DELETE ON analytics_retention_receipts
FOR EACH ROW EXECUTE FUNCTION analytics_reject_mutation_v1();
ALTER TABLE analytics_inbound_references ENABLE ALWAYS TRIGGER analytics_inbound_references_guard;
ALTER TABLE analytics_legal_holds ENABLE ALWAYS TRIGGER analytics_legal_holds_guard;
ALTER TABLE analytics_retention_receipts ENABLE ALWAYS TRIGGER analytics_retention_receipts_immutable;
