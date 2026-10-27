-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_publication_submit_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_release_id uuid,
    p_expected_revision bigint, p_step_up_at timestamptz, p_occurred_at timestamptz
)
RETURNS TABLE (
    release_id uuid, resource_kind text, resource_version_id uuid,
    audience text, lifecycle_state text, cache_policy text,
    manifest_sha256 bytea, requester_id uuid, approver_id uuid,
    row_revision bigint, generation bigint, public_id bytea,
    capability_expires_at timestamptz, created_at timestamptz,
    updated_at timestamptz, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_existing analytics_operations%ROWTYPE; v_manifest bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR p_expected_revision < 1 OR
       p_step_up_at IS NULL OR p_occurred_at IS NULL OR p_step_up_at > p_occurred_at OR
       p_occurred_at - p_step_up_at > interval '10 minutes' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics publication submit input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:publication:submit:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'publication_submit'
      AND operation.idempotency_key = p_idempotency_key FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103', MESSAGE = 'analytics publication submit replay differs';
        END IF;
        RETURN QUERY SELECT result.*, true FROM analytics_publication_get_v1(
            p_scope_schema, v_existing.resource_id) AS result;
        RETURN;
    END IF;
    UPDATE analytics_publication_releases SET lifecycle_state = 'pending_approval',
        requester_step_up_at = p_step_up_at, submitted_at = p_occurred_at,
        row_revision = row_revision + 1, updated_at = p_occurred_at
    WHERE id = p_release_id AND requester_id = p_actor_id AND lifecycle_state = 'draft'
      AND row_revision = p_expected_revision RETURNING manifest_sha256 INTO v_manifest;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics publication submit changed';
    END IF;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at
    ) VALUES (p_operation_id, p_actor_id, 'publication_submit', p_idempotency_key,
        p_request_sha256, 'publication', p_release_id, p_expected_revision + 1,
        v_manifest, p_occurred_at);
    RETURN QUERY SELECT result.*, false FROM analytics_publication_get_v1(
        p_scope_schema, p_release_id) AS result;
END;
$function$;

CREATE FUNCTION analytics_publication_approve_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_release_id uuid,
    p_expected_revision bigint, p_step_up_at timestamptz, p_occurred_at timestamptz
)
RETURNS TABLE (
    release_id uuid, resource_kind text, resource_version_id uuid,
    audience text, lifecycle_state text, cache_policy text,
    manifest_sha256 bytea, requester_id uuid, approver_id uuid,
    row_revision bigint, generation bigint, public_id bytea,
    capability_expires_at timestamptz, created_at timestamptz,
    updated_at timestamptz, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing analytics_operations%ROWTYPE;
    v_requester uuid; v_requester_step_up timestamptz; v_manifest bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR p_expected_revision < 1 OR
       p_step_up_at IS NULL OR p_occurred_at IS NULL OR p_step_up_at > p_occurred_at OR
       p_occurred_at - p_step_up_at > interval '10 minutes' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics publication approval input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:publication:approve:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'publication_approve'
      AND operation.idempotency_key = p_idempotency_key FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103', MESSAGE = 'analytics publication approval replay differs';
        END IF;
        RETURN QUERY SELECT result.*, true FROM analytics_publication_get_v1(
            p_scope_schema, v_existing.resource_id) AS result;
        RETURN;
    END IF;
    SELECT requester_id, requester_step_up_at, manifest_sha256
    INTO v_requester, v_requester_step_up, v_manifest
    FROM analytics_publication_releases WHERE id = p_release_id
      AND lifecycle_state = 'pending_approval' AND approver_id IS NULL
      AND row_revision = p_expected_revision FOR UPDATE;
    IF NOT FOUND OR v_requester = p_actor_id OR v_requester_step_up IS NULL OR
       p_occurred_at - v_requester_step_up > interval '10 minutes' THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics publication approval changed';
    END IF;
    INSERT INTO analytics_publication_approvals VALUES (
        p_release_id, v_requester, v_requester_step_up, p_actor_id, p_step_up_at,
        p_occurred_at, pg_catalog.sha256(p_request_sha256 || v_manifest));
    UPDATE analytics_publication_releases SET approver_id = p_actor_id,
        approved_at = p_occurred_at, row_revision = row_revision + 1,
        updated_at = p_occurred_at WHERE id = p_release_id;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at
    ) VALUES (p_operation_id, p_actor_id, 'publication_approve', p_idempotency_key,
        p_request_sha256, 'publication', p_release_id, p_expected_revision + 1,
        v_manifest, p_occurred_at);
    RETURN QUERY SELECT result.*, false FROM analytics_publication_get_v1(
        p_scope_schema, p_release_id) AS result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_publication_submit_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_publication_approve_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz,timestamptz
) FROM PUBLIC;
