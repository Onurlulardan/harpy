-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_update_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_dashboard_id uuid,
    p_expected_root_revision bigint, p_expected_version_revision bigint,
    p_name text, p_visibility text, p_layout_document bytea,
    p_widgets jsonb, p_occurred_at timestamptz
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
DECLARE v_existing analytics_operations%ROWTYPE; v_version_id uuid; v_manifest bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_expected_root_revision < 1 OR p_expected_version_revision < 1 OR
       pg_catalog.octet_length(p_name) NOT BETWEEN 1 AND 100 OR
       p_visibility NOT IN ('private', 'tenant') OR
       pg_catalog.octet_length(p_layout_document) NOT BETWEEN 2 AND 65536 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard update input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:dashboard:update:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'dashboard_update'
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
    SELECT root.current_draft_version_id INTO v_version_id
    FROM analytics_dashboards AS root
    WHERE root.id = p_dashboard_id AND root.lifecycle_state = 'active'
      AND root.row_revision = p_expected_root_revision FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics dashboard root changed';
    END IF;
    PERFORM 1 FROM analytics_dashboard_versions AS version_row
    WHERE version_row.id = v_version_id AND version_row.dashboard_id = p_dashboard_id
      AND version_row.version_state = 'draft'
      AND version_row.row_revision = p_expected_version_revision FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics dashboard draft changed';
    END IF;
    v_manifest := pg_catalog.sha256(p_layout_document || pg_catalog.convert_to(p_widgets::text, 'UTF8'));
    UPDATE analytics_dashboard_versions SET name = p_name, visibility = p_visibility,
        layout_document = p_layout_document, layout_sha256 = pg_catalog.sha256(p_layout_document),
        manifest_sha256 = v_manifest, row_revision = row_revision + 1,
        updated_by = p_actor_id, updated_at = p_occurred_at WHERE id = v_version_id;
    PERFORM analytics_dashboard_widgets_replace_v1(v_version_id, p_widgets, p_occurred_at);
    UPDATE analytics_dashboards SET row_revision = row_revision + 1,
        updated_by = p_actor_id, updated_at = p_occurred_at WHERE id = p_dashboard_id;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at,
        result_version_id
    ) VALUES (p_operation_id, p_actor_id, 'dashboard_update', p_idempotency_key,
        p_request_sha256, 'dashboard', p_dashboard_id, p_expected_root_revision + 1,
        v_manifest, p_occurred_at, v_version_id);
    RETURN QUERY SELECT result.*, false FROM analytics_dashboard_get_version_v1(
        p_scope_schema, p_dashboard_id, v_version_id) AS result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_update_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,bigint,text,text,bytea,jsonb,timestamptz
) FROM PUBLIC;
