-- +goose Up
-- +goose StatementBegin
DROP FUNCTION analytics_publication_create_v1(
    name,uuid,uuid,text,bytea,uuid,text,uuid,text,text,bytea,timestamptz
);
CREATE FUNCTION analytics_publication_create_v1(
    p_scope_schema name,p_operation_id uuid,p_actor_id uuid,
    p_idempotency_key text,p_request_sha256 bytea,p_release_id uuid,
    p_resource_kind text,p_resource_version_id uuid,p_audience text,
    p_cache_policy text,p_manifest_sha256 bytea,p_occurred_at timestamptz
)
RETURNS TABLE(
    release_id uuid,resource_kind text,resource_version_id uuid,
    audience text,lifecycle_state text,cache_policy text,
    manifest_sha256 bytea,requester_id uuid,approver_id uuid,
    row_revision bigint,generation bigint,public_id bytea,
    capability_expires_at timestamptz,created_at timestamptz,
    updated_at timestamptz,replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing analytics_operations%ROWTYPE;
    v_root_id uuid;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_resource_version_id)
           IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_resource_kind NOT IN ('query','dashboard') OR
       p_audience NOT IN ('tenant','public','capability') OR
       p_cache_policy NOT IN ('none','revision_bound') OR
       pg_catalog.octet_length(p_manifest_sha256) IS DISTINCT FROM 32 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics publication create input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:publication:create:'||p_actor_id::text||':'||
        p_idempotency_key,6100
    ));
    SELECT operation.* INTO v_existing
    FROM analytics_operations AS operation
    WHERE operation.actor_id=p_actor_id
      AND operation.operation_kind='publication_create'
      AND operation.idempotency_key=p_idempotency_key
    FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='P6103',
                MESSAGE='analytics publication replay differs';
        END IF;
        RETURN QUERY SELECT result.*,true
        FROM analytics_publication_get_v1(
            p_scope_schema,v_existing.resource_id
        ) AS result;
        RETURN;
    END IF;
    IF EXISTS (
        SELECT 1 FROM analytics_publication_releases
        WHERE id=p_release_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='analytics publication identity exists';
    END IF;
    IF p_resource_kind='query' THEN
        SELECT version_row.query_id INTO v_root_id
        FROM analytics_saved_query_versions AS version_row
        WHERE version_row.id=p_resource_version_id;
        PERFORM 1 FROM analytics_saved_queries AS root
        WHERE root.id=v_root_id AND root.lifecycle_state='active'
        FOR UPDATE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6102',
                MESSAGE='analytics publication target is inactive';
        END IF;
        PERFORM 1 FROM analytics_saved_query_versions AS version_row
        WHERE version_row.id=p_resource_version_id
          AND version_row.query_id=v_root_id
          AND version_row.version_state IN ('published','needs_revalidation')
        FOR UPDATE;
    ELSE
        SELECT version_row.dashboard_id INTO v_root_id
        FROM analytics_dashboard_versions AS version_row
        WHERE version_row.id=p_resource_version_id;
        PERFORM 1 FROM analytics_dashboards AS root
        WHERE root.id=v_root_id AND root.lifecycle_state='active'
        FOR UPDATE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6102',
                MESSAGE='analytics publication target is inactive';
        END IF;
        PERFORM 1 FROM analytics_dashboard_versions AS version_row
        WHERE version_row.id=p_resource_version_id
          AND version_row.dashboard_id=v_root_id
          AND version_row.version_state='published'
        FOR UPDATE;
    END IF;
    IF NOT FOUND OR v_root_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics publication target is inactive';
    END IF;
    INSERT INTO analytics_publication_releases(
        id,resource_kind,resource_version_id,audience,lifecycle_state,
        cache_policy,manifest_sha256,requester_id,row_revision,generation,
        created_at,updated_at
    ) VALUES (
        p_release_id,p_resource_kind,p_resource_version_id,p_audience,
        'draft',p_cache_policy,p_manifest_sha256,p_actor_id,1,1,
        p_occurred_at,p_occurred_at
    );
    INSERT INTO analytics_operations(
        operation_id,actor_id,operation_kind,idempotency_key,request_sha256,
        resource_kind,resource_id,result_revision,result_sha256,occurred_at,
        result_version_id
    ) VALUES (
        p_operation_id,p_actor_id,'publication_create',p_idempotency_key,
        p_request_sha256,'publication',p_release_id,1,p_manifest_sha256,
        p_occurred_at,p_resource_version_id
    );
    RETURN QUERY SELECT result.*,false
    FROM analytics_publication_get_v1(
        p_scope_schema,p_release_id
    ) AS result;
END;
$function$;
REVOKE ALL ON FUNCTION analytics_publication_create_v1(
    name,uuid,uuid,text,bytea,uuid,text,uuid,text,text,bytea,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_publication_create_v1(
    name,uuid,uuid,text,bytea,uuid,text,uuid,text,text,bytea,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
