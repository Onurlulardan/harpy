-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_widgets_replace_v1(
    p_version_id uuid, p_widgets jsonb, p_occurred_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_count integer;
BEGIN
    IF pg_catalog.jsonb_typeof(p_widgets) IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(p_widgets) > 30 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard widgets are invalid';
    END IF;
    DELETE FROM analytics_dashboard_widgets WHERE dashboard_version_id = p_version_id;
    INSERT INTO analytics_dashboard_widgets (
        dashboard_version_id, widget_id, position, query_version_id,
        parameter_document, parameter_sha256, output_document, output_sha256, created_at
    )
    SELECT p_version_id, (item.value->>'widget_id')::uuid,
           (item.value->>'position')::smallint, (item.value->>'query_version_id')::uuid,
           pg_catalog.decode(item.value->>'parameter_document', 'hex'),
           pg_catalog.sha256(pg_catalog.decode(item.value->>'parameter_document', 'hex')),
           pg_catalog.decode(item.value->>'output_document', 'hex'),
           pg_catalog.sha256(pg_catalog.decode(item.value->>'output_document', 'hex')),
           p_occurred_at
    FROM pg_catalog.jsonb_array_elements(p_widgets) AS item(value);
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count IS DISTINCT FROM pg_catalog.jsonb_array_length(p_widgets) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard widget count differs';
    END IF;
END;
$function$;

CREATE FUNCTION analytics_dashboard_create_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea,
    p_dashboard_id uuid, p_version_id uuid, p_name text, p_visibility text,
    p_layout_document bytea, p_widgets jsonb, p_occurred_at timestamptz
)
RETURNS TABLE (
    dashboard_id uuid, version_id uuid, version_no bigint,
    root_revision bigint, version_revision bigint, lifecycle_state text,
    version_state text, name text, visibility text, layout_document bytea,
    layout_sha256 bytea, manifest_sha256 bytea, widgets jsonb,
    created_at timestamptz, updated_at timestamptz, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_existing analytics_operations%ROWTYPE; v_manifest bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_version_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_name) NOT BETWEEN 1 AND 100 OR
       p_visibility NOT IN ('private', 'tenant') OR
       pg_catalog.octet_length(p_layout_document) NOT BETWEEN 2 AND 65536 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard create input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:dashboard:create:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'dashboard_create'
      AND operation.idempotency_key = p_idempotency_key FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 OR
           v_existing.result_version_id IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103', MESSAGE = 'analytics dashboard replay differs';
        END IF;
        RETURN QUERY SELECT result.*, true FROM analytics_dashboard_get_version_v1(
            p_scope_schema, v_existing.resource_id, v_existing.result_version_id) AS result;
        RETURN;
    END IF;
    IF EXISTS (SELECT 1 FROM analytics_dashboards WHERE id = p_dashboard_id) OR
       EXISTS (SELECT 1 FROM analytics_dashboard_versions WHERE id = p_version_id) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6104', MESSAGE = 'analytics dashboard identity exists';
    END IF;
    v_manifest := pg_catalog.sha256(p_layout_document || pg_catalog.convert_to(p_widgets::text, 'UTF8'));
    SET CONSTRAINTS analytics_dashboards_current_draft_fk DEFERRED;
    INSERT INTO analytics_dashboards VALUES (
        p_dashboard_id, 'active', 1, 1, p_version_id,
        p_actor_id, p_actor_id, p_occurred_at, p_occurred_at);
    INSERT INTO analytics_dashboard_versions (
        id, dashboard_id, version_no, version_state, row_revision, name, visibility,
        layout_document, layout_sha256, manifest_sha256, created_by, updated_by,
        created_at, updated_at
    ) VALUES (p_version_id, p_dashboard_id, 1, 'draft', 1, p_name, p_visibility,
        p_layout_document, pg_catalog.sha256(p_layout_document), v_manifest,
        p_actor_id, p_actor_id, p_occurred_at, p_occurred_at);
    PERFORM analytics_dashboard_widgets_replace_v1(p_version_id, p_widgets, p_occurred_at);
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at,
        result_version_id
    ) VALUES (p_operation_id, p_actor_id, 'dashboard_create', p_idempotency_key,
        p_request_sha256, 'dashboard', p_dashboard_id, 1, v_manifest, p_occurred_at, p_version_id);
    RETURN QUERY SELECT result.*, false FROM analytics_dashboard_get_version_v1(
        p_scope_schema, p_dashboard_id, p_version_id) AS result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_widgets_replace_v1(uuid,jsonb,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_create_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,text,text,bytea,jsonb,timestamptz
) FROM PUBLIC;
