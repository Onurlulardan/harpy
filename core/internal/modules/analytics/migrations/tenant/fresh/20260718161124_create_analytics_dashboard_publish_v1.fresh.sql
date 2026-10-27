-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_publish_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_dashboard_id uuid,
    p_draft_version_id uuid, p_next_draft_version_id uuid,
    p_expected_root_revision bigint, p_expected_draft_revision bigint,
    p_manifest_sha256 bytea, p_occurred_at timestamptz
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
DECLARE
    v_existing analytics_operations%ROWTYPE;
    v_version analytics_dashboard_versions%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_draft_version_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_next_draft_version_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_expected_root_revision < 1 OR p_expected_draft_revision < 1 OR
       pg_catalog.octet_length(p_manifest_sha256) IS DISTINCT FROM 32 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard publish input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:dashboard:publish:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'dashboard_publish'
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
    SELECT version_row.* INTO v_version FROM analytics_dashboards AS root
    JOIN analytics_dashboard_versions AS version_row
      ON version_row.id = root.current_draft_version_id AND version_row.dashboard_id = root.id
    WHERE root.id = p_dashboard_id AND root.lifecycle_state = 'active'
      AND root.row_revision = p_expected_root_revision
      AND version_row.id = p_draft_version_id AND version_row.version_state = 'draft'
      AND version_row.row_revision = p_expected_draft_revision FOR UPDATE OF root, version_row;
    IF NOT FOUND OR NOT EXISTS (SELECT 1 FROM analytics_dashboard_widgets
        WHERE dashboard_version_id = p_draft_version_id) THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics dashboard draft changed';
    END IF;
    IF EXISTS (SELECT 1 FROM analytics_dashboard_versions WHERE id = p_next_draft_version_id) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6104', MESSAGE = 'analytics dashboard version exists';
    END IF;
    UPDATE analytics_dashboard_versions SET version_state = 'published',
        row_revision = row_revision + 1, manifest_sha256 = p_manifest_sha256,
        updated_by = p_actor_id, updated_at = p_occurred_at,
        published_by = p_actor_id, published_at = p_occurred_at
    WHERE id = p_draft_version_id;
    SET CONSTRAINTS analytics_dashboards_current_draft_fk DEFERRED;
    INSERT INTO analytics_dashboard_versions (
        id, dashboard_id, version_no, version_state, row_revision, name, visibility,
        layout_document, layout_sha256, manifest_sha256, created_by, updated_by,
        created_at, updated_at
    ) VALUES (p_next_draft_version_id, p_dashboard_id, v_version.version_no + 1,
        'draft', 1, v_version.name, v_version.visibility, v_version.layout_document,
        v_version.layout_sha256, p_manifest_sha256, p_actor_id, p_actor_id,
        p_occurred_at, p_occurred_at);
    INSERT INTO analytics_dashboard_widgets SELECT p_next_draft_version_id,
        widget_id, position, query_version_id, parameter_document, parameter_sha256,
        output_document, output_sha256, p_occurred_at FROM analytics_dashboard_widgets
        WHERE dashboard_version_id = p_draft_version_id;
    UPDATE analytics_dashboards SET current_draft_version_id = p_next_draft_version_id,
        row_revision = row_revision + 1, updated_by = p_actor_id,
        updated_at = p_occurred_at WHERE id = p_dashboard_id;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at,
        result_version_id
    ) VALUES (p_operation_id, p_actor_id, 'dashboard_publish', p_idempotency_key,
        p_request_sha256, 'dashboard', p_dashboard_id, p_expected_root_revision + 1,
        p_manifest_sha256, p_occurred_at, p_draft_version_id);
    RETURN QUERY SELECT result.*, false FROM analytics_dashboard_get_version_v1(
        p_scope_schema, p_dashboard_id, p_draft_version_id) AS result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_publish_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,bigint,bigint,bytea,timestamptz
) FROM PUBLIC;
