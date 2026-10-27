-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_set_default_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_principal_id uuid,
    p_dashboard_id uuid, p_surface text, p_occurred_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_existing analytics_operations%ROWTYPE; v_revision bigint;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_principal_id IS NULL OR p_principal_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_surface IS NULL OR p_surface OPERATOR(pg_catalog.!~) '^[a-z][a-z0-9_.-]{0,63}$' OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard default input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:dashboard:default:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'dashboard_set_default'
      AND operation.idempotency_key = p_idempotency_key FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103', MESSAGE = 'analytics dashboard default replay differs';
        END IF;
        RETURN true;
    END IF;
    PERFORM 1 FROM analytics_dashboards AS root
    JOIN analytics_dashboard_versions AS version_row ON version_row.dashboard_id = root.id
    WHERE root.id = p_dashboard_id AND root.lifecycle_state = 'active'
      AND version_row.version_state = 'published';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'published analytics dashboard was not found';
    END IF;
    INSERT INTO analytics_dashboard_preferences AS preference (
        principal_id, surface, dashboard_id, row_revision, updated_by, updated_at
    ) VALUES (p_principal_id, p_surface, p_dashboard_id, 1, p_actor_id, p_occurred_at)
    ON CONFLICT (principal_id, surface) DO UPDATE
    SET dashboard_id = EXCLUDED.dashboard_id, row_revision = preference.row_revision + 1,
        updated_by = EXCLUDED.updated_by, updated_at = EXCLUDED.updated_at
    RETURNING row_revision INTO v_revision;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at
    ) VALUES (p_operation_id, p_actor_id, 'dashboard_set_default', p_idempotency_key,
        p_request_sha256, 'dashboard', p_dashboard_id, v_revision,
        p_request_sha256, p_occurred_at);
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_set_default_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,text,timestamptz
) FROM PUBLIC;
