-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_forms_reference_claim_v1(
    p_scope_schema name,p_reference_id uuid,p_operation_id uuid,
    p_source_kind text,p_source_id uuid,p_source_version_id uuid,
    p_component_id uuid,p_target_definition_id uuid,p_target_version_id uuid,
    p_actor_id uuid,p_occurred_at timestamptz
)
RETURNS SETOF analytics_forms_reference_bindings
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_existing analytics_forms_reference_bindings%ROWTYPE;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    IF pg_catalog.uuid_extract_version(p_reference_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_reference_id=p_operation_id OR
       p_source_kind NOT IN ('query','catalog_query') OR
       pg_catalog.uuid_extract_version(p_source_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_source_version_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_target_definition_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_target_version_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at OR
       (p_source_kind='query' AND p_component_id IS NOT NULL) OR
       (p_source_kind='catalog_query' AND
        pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Forms reference claim is invalid';
    END IF;
    PERFORM 1
    FROM analytics_saved_query_versions AS version_row
    JOIN analytics_query_dependencies AS dependency
      ON dependency.query_version_id=version_row.id
     AND dependency.owner_domain='forms'
     AND dependency.resource_kind='form'
     AND dependency.resource_id=p_target_definition_id
     AND dependency.resource_version_id=p_target_version_id
    WHERE version_row.query_id=p_source_id
      AND version_row.id=p_source_version_id
      AND version_row.version_state IN ('published','retired','needs_revalidation');
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Forms reference dependency is unavailable';
    END IF;
    IF p_source_kind='catalog_query' AND NOT EXISTS (
        SELECT 1 FROM analytics_catalog_version_projections AS projection
        WHERE projection.component_id=p_component_id
          AND projection.analytics_kind='endpoint'
          AND projection.operational_state='ready'
          AND projection.operational_resource_id=p_source_id
          AND projection.published_version_id=p_source_version_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog Forms reference source differs';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_source_id::text||':'||p_source_version_id::text||':'||
        p_target_definition_id::text||':'||p_target_version_id::text,145401
    ));
    SELECT * INTO v_existing
    FROM analytics_forms_reference_bindings AS binding
    WHERE binding.source_id=p_source_id
      AND binding.source_version_id=p_source_version_id
      AND binding.target_definition_id=p_target_definition_id
      AND binding.target_version_id=p_target_version_id
      AND binding.usage='READ'
    FOR UPDATE;
    IF FOUND THEN
        IF v_existing.source_kind<>p_source_kind OR
           v_existing.component_id IS DISTINCT FROM p_component_id OR
           v_existing.claimed_by<>p_actor_id THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Analytics Forms reference claim differs';
        END IF;
        RETURN NEXT v_existing;
        RETURN;
    END IF;
    INSERT INTO analytics_forms_reference_bindings(
        reference_id,register_operation_id,source_kind,source_id,
        source_version_id,component_id,target_definition_id,
        target_version_id,claimed_by,claimed_at
    ) VALUES (
        p_reference_id,p_operation_id,p_source_kind,p_source_id,
        p_source_version_id,p_component_id,p_target_definition_id,
        p_target_version_id,p_actor_id,p_occurred_at
    ) RETURNING * INTO v_existing;
    RETURN NEXT v_existing;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_forms_reference_claim_v1(
    name,uuid,uuid,text,uuid,uuid,uuid,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
