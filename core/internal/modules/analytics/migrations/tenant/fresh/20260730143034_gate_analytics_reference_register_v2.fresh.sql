-- +goose Up
DROP FUNCTION analytics_reference_register_v1(
    name,uuid,uuid,text,uuid,text,text,uuid,uuid,text,timestamptz
);
-- +goose StatementBegin
CREATE FUNCTION analytics_reference_register_v1(
    p_scope_schema name,p_reference_id uuid,p_actor_id uuid,
    p_target_kind text,p_target_id uuid,p_source_module text,
    p_source_kind text,p_source_id uuid,p_source_version_id uuid,
    p_usage text,p_created_at timestamptz
)
RETURNS TABLE(reference_id uuid,state text,created_at timestamptz,replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing analytics_inbound_references%ROWTYPE;
    v_target_ready boolean;
BEGIN
    IF pg_catalog.uuid_extract_version(p_reference_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       p_target_kind NOT IN (
           'query_version','dashboard_version','publication_release'
       ) OR pg_catalog.uuid_extract_version(p_target_id) IS DISTINCT FROM 7 OR
       p_source_module!~'^[a-z][a-z0-9_]{0,63}$' OR
       p_source_kind!~'^[a-z][a-z0-9_.-]{0,63}$' OR
       pg_catalog.uuid_extract_version(p_source_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_source_version_id)
           IS DISTINCT FROM 7 OR
       p_usage NOT IN ('read','write','schema','lifecycle') OR
       p_created_at IS NULL OR NOT pg_catalog.isfinite(p_created_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics reference registration is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:reference:'||p_target_kind||':'||p_target_id::text,6100
    ));
    SELECT * INTO v_existing
    FROM analytics_inbound_references AS reference_row
    WHERE reference_row.reference_id=p_reference_id FOR UPDATE;
    IF FOUND THEN
        IF v_existing.target_kind IS DISTINCT FROM p_target_kind OR
           v_existing.target_id IS DISTINCT FROM p_target_id OR
           v_existing.source_module IS DISTINCT FROM p_source_module OR
           v_existing.source_kind IS DISTINCT FROM p_source_kind OR
           v_existing.source_id IS DISTINCT FROM p_source_id OR
           v_existing.source_version_id IS DISTINCT FROM p_source_version_id OR
           v_existing.usage IS DISTINCT FROM p_usage THEN
            RAISE EXCEPTION USING ERRCODE='P6103',
                MESSAGE='analytics reference identity was reused';
        END IF;
        RETURN QUERY SELECT v_existing.reference_id,v_existing.state,
            v_existing.created_at,true;
        RETURN;
    END IF;
    v_target_ready:=false;
    IF p_target_kind='query_version' THEN
        SELECT true INTO v_target_ready
        FROM analytics_saved_query_versions AS version_row
        JOIN analytics_saved_queries AS root
          ON root.id=version_row.query_id
         AND root.lifecycle_state='active'
        WHERE version_row.id=p_target_id
          AND version_row.version_state IN ('published','needs_revalidation')
        FOR UPDATE OF root,version_row;
    ELSIF p_target_kind='dashboard_version' THEN
        SELECT true INTO v_target_ready
        FROM analytics_dashboard_versions AS version_row
        JOIN analytics_dashboards AS root
          ON root.id=version_row.dashboard_id
         AND root.lifecycle_state='active'
        WHERE version_row.id=p_target_id
          AND version_row.version_state='published'
        FOR UPDATE OF root,version_row;
    ELSE
        SELECT true INTO v_target_ready
        FROM analytics_publication_releases AS release
        WHERE release.id=p_target_id AND release.lifecycle_state='active'
        FOR UPDATE;
    END IF;
    IF NOT COALESCE(v_target_ready,false) THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics reference target is not active';
    END IF;
    BEGIN
        INSERT INTO analytics_inbound_references(
            reference_id,target_kind,target_id,source_module,source_kind,
            source_id,source_version_id,usage,state,created_at
        ) VALUES (
            p_reference_id,p_target_kind,p_target_id,p_source_module,
            p_source_kind,p_source_id,p_source_version_id,p_usage,
            'active',p_created_at
        );
    EXCEPTION WHEN unique_violation THEN
        RAISE EXCEPTION USING ERRCODE='P6103',
            MESSAGE='analytics reference already exists';
    END;
    RETURN QUERY SELECT p_reference_id,'active'::text,p_created_at,false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_reference_register_v1(
    name,uuid,uuid,text,uuid,text,text,uuid,uuid,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_reference_register_v1(
    name,uuid,uuid,text,uuid,text,text,uuid,uuid,text,timestamptz
) TO harpy_runtime;
