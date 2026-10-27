-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_delete_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_dashboard_id uuid,
    p_expected_root_revision bigint, p_occurred_at timestamptz
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
       p_expected_root_revision < 1 OR p_occurred_at IS NULL OR
       NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard delete input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:dashboard:delete:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'dashboard_delete'
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
    SELECT root.current_draft_version_id, version_row.manifest_sha256
    INTO v_version_id, v_manifest FROM analytics_dashboards AS root
    JOIN analytics_dashboard_versions AS version_row
      ON version_row.id = root.current_draft_version_id
    WHERE root.id = p_dashboard_id AND root.lifecycle_state = 'active'
      AND root.row_revision = p_expected_root_revision FOR UPDATE OF root, version_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics dashboard root changed';
    END IF;
    UPDATE analytics_dashboards SET lifecycle_state = 'retired',
        row_revision = row_revision + 1, reference_epoch = reference_epoch + 1,
        updated_by = p_actor_id, updated_at = p_occurred_at WHERE id = p_dashboard_id;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at,
        result_version_id
    ) VALUES (p_operation_id, p_actor_id, 'dashboard_delete', p_idempotency_key,
        p_request_sha256, 'dashboard', p_dashboard_id, p_expected_root_revision + 1,
        v_manifest, p_occurred_at, v_version_id);
    RETURN QUERY SELECT result.*, false FROM analytics_dashboard_get_version_v1(
        p_scope_schema, p_dashboard_id, v_version_id) AS result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_delete_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz
) FROM PUBLIC;
