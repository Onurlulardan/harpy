-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_publication_activate_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_release_id uuid,
    p_expected_revision bigint, p_step_up_at timestamptz, p_public_id bytea,
    p_capability_sha256 bytea, p_capability_expires_at timestamptz,
    p_occurred_at timestamptz
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
    v_existing analytics_operations%ROWTYPE; v_release analytics_publication_releases%ROWTYPE;
    v_tenant_id uuid;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR p_expected_revision < 1 OR
       p_step_up_at IS NULL OR p_occurred_at IS NULL OR p_step_up_at > p_occurred_at OR
       p_occurred_at - p_step_up_at > interval '10 minutes' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics publication activation input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:publication:activate:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'publication_activate'
      AND operation.idempotency_key = p_idempotency_key FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103', MESSAGE = 'analytics publication activation replay differs';
        END IF;
        RETURN QUERY SELECT result.*, true FROM analytics_publication_get_v1(
            p_scope_schema, v_existing.resource_id) AS result;
        RETURN;
    END IF;
    SELECT release.* INTO v_release FROM analytics_publication_releases AS release
    WHERE release.id = p_release_id AND release.lifecycle_state = 'pending_approval'
      AND release.row_revision = p_expected_revision FOR UPDATE;
    IF NOT FOUND OR (v_release.audience <> 'tenant' AND v_release.approver_id IS NULL) OR
       (v_release.audience = 'tenant' AND (p_public_id IS NOT NULL OR p_capability_sha256 IS NOT NULL
          OR p_capability_expires_at IS NOT NULL)) OR
       (v_release.audience = 'public' AND (pg_catalog.octet_length(p_public_id) IS DISTINCT FROM 16
          OR p_capability_sha256 IS NOT NULL OR p_capability_expires_at IS NOT NULL)) OR
       (v_release.audience = 'capability' AND (pg_catalog.octet_length(p_public_id) IS DISTINCT FROM 16
          OR pg_catalog.octet_length(p_capability_sha256) IS DISTINCT FROM 32
          OR p_capability_expires_at IS NULL OR p_capability_expires_at <= p_occurred_at
          OR p_capability_expires_at - p_occurred_at > interval '30 days')) THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics publication activation changed';
    END IF;
    UPDATE analytics_publication_releases SET lifecycle_state = 'active',
        public_id = p_public_id, capability_expires_at = p_capability_expires_at,
        activated_at = p_occurred_at, row_revision = row_revision + 1,
        updated_at = p_occurred_at WHERE id = p_release_id;
    IF v_release.audience = 'capability' THEN
        INSERT INTO analytics_capability_tokens VALUES (
            p_release_id, p_capability_sha256, p_occurred_at, p_capability_expires_at, NULL);
    END IF;
    IF v_release.audience <> 'tenant' THEN
        SELECT tenant_id INTO STRICT v_tenant_id FROM tenancy_schema_identity
        WHERE singleton AND schema_name = pg_catalog.current_schema();
        PERFORM platform.analytics_public_activate_v1(
            v_tenant_id,p_public_id,p_release_id,v_release.resource_kind,
            v_release.resource_version_id,v_release.audience,v_release.generation,
            p_capability_sha256,p_capability_expires_at,p_occurred_at);
    END IF;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at
    ) VALUES (p_operation_id, p_actor_id, 'publication_activate', p_idempotency_key,
        p_request_sha256, 'publication', p_release_id, p_expected_revision + 1,
        v_release.manifest_sha256, p_occurred_at);
    RETURN QUERY SELECT result.*, false FROM analytics_publication_get_v1(
        p_scope_schema, p_release_id) AS result;
END;
$function$;

CREATE FUNCTION analytics_publication_revoke_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_release_id uuid,
    p_expected_revision bigint, p_occurred_at timestamptz
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
    v_release analytics_publication_releases%ROWTYPE;
    v_tenant_id uuid;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR p_expected_revision < 1 OR
       p_occurred_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics publication revoke input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:publication:revoke:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id AND operation.operation_kind = 'publication_revoke'
      AND operation.idempotency_key = p_idempotency_key FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103', MESSAGE = 'analytics publication revoke replay differs';
        END IF;
        RETURN QUERY SELECT result.*, true FROM analytics_publication_get_v1(
            p_scope_schema, v_existing.resource_id) AS result; RETURN;
    END IF;
    SELECT release.* INTO v_release FROM analytics_publication_releases AS release
    WHERE release.id = p_release_id AND release.lifecycle_state = 'active'
      AND release.row_revision = p_expected_revision FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'analytics publication revoke changed';
    END IF;
    UPDATE analytics_publication_releases SET lifecycle_state = 'revoked',
        generation = generation + 1, revoked_at = p_occurred_at,
        row_revision = row_revision + 1, updated_at = p_occurred_at WHERE id = p_release_id;
    UPDATE analytics_capability_tokens SET revoked_at = p_occurred_at WHERE release_id = p_release_id;
    IF v_release.audience <> 'tenant' THEN
        SELECT tenant_id INTO STRICT v_tenant_id FROM tenancy_schema_identity
        WHERE singleton AND schema_name = pg_catalog.current_schema();
        PERFORM platform.analytics_public_revoke_v1(
            v_tenant_id,p_release_id,v_release.generation,p_occurred_at);
    END IF;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key, request_sha256,
        resource_kind, resource_id, result_revision, result_sha256, occurred_at
    ) VALUES (p_operation_id, p_actor_id, 'publication_revoke', p_idempotency_key,
        p_request_sha256, 'publication', p_release_id, p_expected_revision + 1,
        v_release.manifest_sha256, p_occurred_at);
    RETURN QUERY SELECT result.*, false FROM analytics_publication_get_v1(
        p_scope_schema, p_release_id) AS result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_publication_activate_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz,bytea,bytea,timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_publication_revoke_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz
) FROM PUBLIC;
