-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_grant_acl_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_dashboard_id uuid,
    p_principal_kind text, p_principal_id uuid, p_access_level text,
    p_expected_revision bigint, p_occurred_at timestamptz
)
RETURNS TABLE (
    dashboard_id uuid, principal_kind text, principal_id uuid,
    access_level text, row_revision bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_existing analytics_operations%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_principal_kind NOT IN ('user', 'role') OR
       p_principal_id IS NULL OR p_principal_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_access_level NOT IN ('read', 'edit', 'publish') OR p_expected_revision < 0 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard ACL input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:dashboard:grant-acl:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'dashboard_grant_acl'
      AND operation.idempotency_key = p_idempotency_key FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103', MESSAGE = 'analytics dashboard ACL replay differs';
        END IF;
        RETURN QUERY SELECT acl.dashboard_id, acl.principal_kind, acl.principal_id,
            acl.access_level, acl.row_revision, true FROM analytics_dashboard_acls AS acl
            WHERE acl.dashboard_id = p_dashboard_id AND acl.principal_kind = p_principal_kind
              AND acl.principal_id = p_principal_id;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103', MESSAGE = 'analytics dashboard ACL replay expired';
        END IF;
        RETURN;
    END IF;
    PERFORM 1 FROM analytics_dashboards WHERE id = p_dashboard_id AND lifecycle_state = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics dashboard was not found';
    END IF;
    INSERT INTO analytics_dashboard_acls AS acl (
        dashboard_id, principal_kind, principal_id, access_level,
        row_revision, granted_by, granted_at
    ) VALUES (p_dashboard_id, p_principal_kind, p_principal_id, p_access_level,
        1, p_actor_id, p_occurred_at)
    ON CONFLICT (dashboard_id, principal_kind, principal_id) DO UPDATE
    SET access_level = EXCLUDED.access_level, row_revision = acl.row_revision + 1,
        granted_by = EXCLUDED.granted_by, granted_at = EXCLUDED.granted_at
    WHERE p_expected_revision > 0 AND acl.row_revision = p_expected_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics dashboard ACL changed';
    END IF;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at
    ) VALUES (p_operation_id, p_actor_id, 'dashboard_grant_acl', p_idempotency_key,
        p_request_sha256, 'dashboard', p_dashboard_id, p_expected_revision + 1,
        p_request_sha256, p_occurred_at);
    RETURN QUERY SELECT acl.dashboard_id, acl.principal_kind, acl.principal_id,
        acl.access_level, acl.row_revision, false FROM analytics_dashboard_acls AS acl
        WHERE acl.dashboard_id = p_dashboard_id AND acl.principal_kind = p_principal_kind
          AND acl.principal_id = p_principal_id;
END;
$function$;

CREATE FUNCTION analytics_dashboard_revoke_acl_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_dashboard_id uuid,
    p_principal_kind text, p_principal_id uuid, p_expected_revision bigint,
    p_occurred_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_existing analytics_operations%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_principal_kind NOT IN ('user', 'role') OR p_principal_id IS NULL OR
       p_principal_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_expected_revision < 1 OR p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard ACL revoke input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:dashboard:revoke-acl:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'dashboard_revoke_acl'
      AND operation.idempotency_key = p_idempotency_key FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103', MESSAGE = 'analytics dashboard ACL replay differs';
        END IF;
        RETURN true;
    END IF;
    DELETE FROM analytics_dashboard_acls WHERE dashboard_id = p_dashboard_id
      AND principal_kind = p_principal_kind AND principal_id = p_principal_id
      AND row_revision = p_expected_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics dashboard ACL changed';
    END IF;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at
    ) VALUES (p_operation_id, p_actor_id, 'dashboard_revoke_acl', p_idempotency_key,
        p_request_sha256, 'dashboard', p_dashboard_id, p_expected_revision + 1,
        p_request_sha256, p_occurred_at);
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_grant_acl_v1(
    name,uuid,uuid,text,bytea,uuid,text,uuid,text,bigint,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_revoke_acl_v1(
    name,uuid,uuid,text,bytea,uuid,text,uuid,bigint,timestamptz
) FROM PUBLIC;
