-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_activate_v2(
    p_scope_schema name,p_projection_receipt_id uuid,p_operation_id uuid,
    p_visible boolean,p_completed_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_projection analytics_catalog_version_projections%ROWTYPE;
    v_candidate analytics_catalog_component_candidates%ROWTYPE;
    v_attempt analytics_catalog_projection_attempts%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_projection_receipt_id)
           IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_visible IS NULL OR p_completed_at IS NULL OR
       NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog v2 activation input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT candidate.* INTO v_candidate
    FROM analytics_catalog_component_candidates AS candidate
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=candidate.projection_receipt_id
     AND projection.component_id=candidate.component_id
     AND projection.package_id=candidate.package_id
     AND projection.package_version=candidate.package_version
     AND projection.release_id=candidate.release_id
     AND projection.owner_generation=candidate.owner_generation
     AND projection.content_sha256=candidate.content_sha256
    JOIN analytics_catalog_projection_attempts AS attempt
      ON attempt.receipt_id=candidate.attempt_receipt_id
     AND attempt.projection_receipt_id=projection.receipt_id
    WHERE projection.receipt_id=p_projection_receipt_id
      AND candidate.staged_operation_id=p_operation_id
    FOR UPDATE OF projection,candidate,attempt;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog v2 activation candidate differs';
    END IF;
    SELECT * INTO v_projection
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.receipt_id=v_candidate.projection_receipt_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog v2 activation candidate differs';
    END IF;
    SELECT * INTO v_attempt
    FROM analytics_catalog_projection_attempts AS attempt
    WHERE attempt.receipt_id=v_candidate.attempt_receipt_id
      AND attempt.projection_receipt_id=v_projection.receipt_id;
    IF NOT FOUND OR v_projection.operational_state<>'ready' OR
       v_projection.blocker_code IS NOT NULL OR
       v_projection.ready_at IS NULL OR v_attempt.ready_at IS NULL OR
       v_candidate.prepared_at>p_completed_at OR
       v_attempt.updated_at>p_completed_at OR
       pg_catalog.uuid_extract_version(
           v_projection.operational_resource_id
       ) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(
           v_projection.published_version_id
       ) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(
           v_projection.current_draft_version_id
       ) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog v2 ready candidate is unavailable';
    END IF;
    PERFORM analytics_catalog_operational_activate_v2(
        p_scope_schema,v_projection.receipt_id,p_operation_id,
        p_visible,p_completed_at
    );
    INSERT INTO analytics_catalog_definitions(
        component_id,analytics_kind,module_slug,slug,name,description,
        legacy_hash,definition_json,definition_sha256,document_bytes,
        content_sha256,semantic_sha256,runtime_plan_bytes,
        runtime_plan_sha256,package_id,package_slug,package_version,
        release_id,operational_state,blocker_code,operational_resource_id,
        published_version_id,current_draft_version_id,installed_at,updated_at
    ) VALUES (
        v_projection.component_id,v_projection.analytics_kind,
        v_projection.module_slug,v_projection.semantic_identity,
        v_projection.name,v_projection.description,v_projection.legacy_hash,
        v_projection.definition_json,v_projection.definition_sha256,
        v_projection.document_bytes,v_projection.content_sha256,
        v_projection.semantic_sha256,v_projection.runtime_plan_bytes,
        v_projection.runtime_plan_sha256,v_projection.package_id,
        v_projection.package_slug,v_projection.package_version,
        v_projection.release_id,'ready',NULL,
        v_projection.operational_resource_id,
        v_projection.published_version_id,
        v_projection.current_draft_version_id,
        v_projection.created_at,p_completed_at
    ) ON CONFLICT (component_id) DO UPDATE SET
        analytics_kind=EXCLUDED.analytics_kind,
        module_slug=EXCLUDED.module_slug,slug=EXCLUDED.slug,
        name=EXCLUDED.name,description=EXCLUDED.description,
        legacy_hash=EXCLUDED.legacy_hash,
        definition_json=EXCLUDED.definition_json,
        definition_sha256=EXCLUDED.definition_sha256,
        document_bytes=EXCLUDED.document_bytes,
        content_sha256=EXCLUDED.content_sha256,
        semantic_sha256=EXCLUDED.semantic_sha256,
        runtime_plan_bytes=EXCLUDED.runtime_plan_bytes,
        runtime_plan_sha256=EXCLUDED.runtime_plan_sha256,
        package_id=EXCLUDED.package_id,package_slug=EXCLUDED.package_slug,
        package_version=EXCLUDED.package_version,
        release_id=EXCLUDED.release_id,
        operational_state='ready',blocker_code=NULL,
        operational_resource_id=EXCLUDED.operational_resource_id,
        published_version_id=EXCLUDED.published_version_id,
        current_draft_version_id=EXCLUDED.current_draft_version_id,
        installed_at=EXCLUDED.installed_at,updated_at=EXCLUDED.updated_at;
    UPDATE analytics_catalog_projection_activations AS activation
    SET superseded_at=p_completed_at
    WHERE activation.component_id=v_projection.component_id
      AND activation.superseded_at IS NULL
      AND activation.projection_receipt_id<>v_projection.receipt_id;
    INSERT INTO analytics_catalog_projection_activations(
        projection_receipt_id,attempt_receipt_id,component_id,
        package_id,package_version,release_id,content_sha256,activated_at
    ) VALUES (
        v_projection.receipt_id,v_attempt.receipt_id,
        v_projection.component_id,v_projection.package_id,
        v_projection.package_version,v_projection.release_id,
        v_projection.content_sha256,p_completed_at
    );
    INSERT INTO analytics_catalog_component_heads(
        component_id,projection_receipt_id,package_id,package_version,
        release_id,content_sha256,activated_at
    ) VALUES (
        v_projection.component_id,v_projection.receipt_id,
        v_projection.package_id,v_projection.package_version,
        v_projection.release_id,v_projection.content_sha256,p_completed_at
    ) ON CONFLICT ON CONSTRAINT analytics_catalog_component_heads_pkey
      DO UPDATE SET
        projection_receipt_id=EXCLUDED.projection_receipt_id,
        package_id=EXCLUDED.package_id,
        package_version=EXCLUDED.package_version,
        release_id=EXCLUDED.release_id,
        content_sha256=EXCLUDED.content_sha256,
        activated_at=EXCLUDED.activated_at;
    DELETE FROM analytics_catalog_component_candidates AS candidate
    WHERE candidate.component_id=v_projection.component_id
      AND candidate.projection_receipt_id=v_projection.receipt_id
      AND candidate.attempt_receipt_id=v_attempt.receipt_id
      AND candidate.staged_operation_id=p_operation_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog activated candidate cleanup differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_projection_activate_v2(
    name,uuid,uuid,boolean,timestamptz
) FROM PUBLIC;
