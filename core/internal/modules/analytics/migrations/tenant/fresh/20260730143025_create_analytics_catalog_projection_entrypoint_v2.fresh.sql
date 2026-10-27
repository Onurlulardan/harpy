-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_read_v2(p_receipt_id uuid)
RETURNS TABLE(
    receipt_id uuid,package_id uuid,package_slug text,
    package_version text,release_id uuid,component_id uuid,
    owner_generation bigint,contract text,content_sha256 text,
    component_kind text,module_slug text,semantic_identity text,
    semantic_sha256 bytea,item_count integer,auxiliary_count integer,
    document_bytes bytea,idempotency_key text,request_sha256 bytea,
    lifecycle_state text,ready_at timestamptz,
    created_at timestamptz,updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN QUERY
    SELECT attempt.receipt_id,projection.package_id,
        projection.package_slug,projection.package_version,
        projection.release_id,projection.component_id,
        projection.owner_generation,projection.contract,
        projection.content_sha256,projection.component_kind,
        projection.module_slug,projection.semantic_identity,
        projection.semantic_sha256,projection.item_count,
        projection.auxiliary_count,projection.document_bytes,
        attempt.idempotency_key,attempt.request_sha256,
        CASE
          WHEN candidate.attempt_receipt_id=attempt.receipt_id
           AND head.component_id IS NULL
           AND component.lifecycle_state='AVAILABLE'
            THEN 'AVAILABLE'
          ELSE 'STAGED'
        END,
        attempt.ready_at,attempt.created_at,attempt.updated_at
    FROM analytics_catalog_projection_attempts AS attempt
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=attempt.projection_receipt_id
    JOIN analytics_catalog_components AS component
      ON component.component_id=projection.component_id
    LEFT JOIN analytics_catalog_component_candidates AS candidate
      ON candidate.component_id=projection.component_id
     AND candidate.projection_receipt_id=projection.receipt_id
    LEFT JOIN analytics_catalog_component_heads AS head
      ON head.component_id=projection.component_id
    WHERE attempt.receipt_id=p_receipt_id
      AND component.lifecycle_state<>'PURGED'
      AND (
        (projection.operational_state='pending'
         AND projection.blocker_code IS NOT NULL
         AND projection.ready_at IS NULL
         AND attempt.ready_at IS NULL)
        OR
        (projection.operational_state='ready'
         AND projection.blocker_code IS NULL
         AND projection.operational_resource_id IS NOT NULL
         AND projection.published_version_id IS NOT NULL
         AND projection.current_draft_version_id IS NOT NULL
         AND projection.ready_at IS NOT NULL
         AND attempt.ready_at IS NOT NULL)
      );
END;
$function$;
REVOKE ALL ON FUNCTION analytics_catalog_projection_read_v2(uuid) FROM PUBLIC;

CREATE FUNCTION analytics_catalog_project_v2(
    p_scope_schema name,p_receipt_id uuid,
    p_package_id uuid,p_package_slug text,p_package_version text,
    p_release_id uuid,p_component_id uuid,p_owner_generation bigint,
    p_contract text,p_content_sha256 text,p_component_kind text,
    p_module_slug text,p_semantic_identity text,p_semantic_sha256 bytea,
    p_item_count integer,p_auxiliary_count integer,p_document_bytes bytea,
    p_runtime_plan_bytes bytea,p_runtime_plan_sha256 bytea,p_actor_id uuid,
    p_idempotency_key text,p_request_sha256 bytea,p_occurred_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid,package_id uuid,package_slug text,
    package_version text,release_id uuid,component_id uuid,
    owner_generation bigint,contract text,content_sha256 text,
    component_kind text,module_slug text,semantic_identity text,
    semantic_sha256 bytea,item_count integer,auxiliary_count integer,
    document_bytes bytea,idempotency_key text,request_sha256 bytea,
    lifecycle_state text,ready_at timestamptz,
    created_at timestamptz,updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_projection_receipt_id uuid;
    v_attempt_receipt_id uuid;
BEGIN
    v_projection_receipt_id:=analytics_catalog_projection_begin_v2(
        p_scope_schema,p_receipt_id,p_package_id,p_package_slug,
        p_package_version,p_release_id,p_component_id,p_owner_generation,
        p_contract,p_content_sha256,p_component_kind,p_module_slug,
        p_semantic_identity,p_semantic_sha256,p_item_count,p_auxiliary_count,
        p_document_bytes,p_runtime_plan_bytes,p_runtime_plan_sha256,p_actor_id,
        p_idempotency_key,p_request_sha256,p_occurred_at
    );
    v_attempt_receipt_id:=analytics_catalog_projection_attempt_v2(
        p_scope_schema,p_receipt_id,v_projection_receipt_id,p_actor_id,
        p_idempotency_key,p_request_sha256,p_occurred_at
    );
    PERFORM analytics_catalog_projection_materialize_v2(
        p_scope_schema,v_attempt_receipt_id,p_occurred_at
    );
    RETURN QUERY
    SELECT snapshot.receipt_id,snapshot.package_id,snapshot.package_slug,
        snapshot.package_version,snapshot.release_id,snapshot.component_id,
        snapshot.owner_generation,snapshot.contract,snapshot.content_sha256,
        snapshot.component_kind,snapshot.module_slug,
        snapshot.semantic_identity,snapshot.semantic_sha256,
        snapshot.item_count,snapshot.auxiliary_count,snapshot.document_bytes,
        snapshot.idempotency_key,snapshot.request_sha256,
        snapshot.lifecycle_state,snapshot.ready_at,
        snapshot.created_at,snapshot.updated_at
    FROM analytics_catalog_projection_read_v2(
        v_attempt_receipt_id
    ) AS snapshot;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog v2 projection readback differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_project_v2(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) FROM PUBLIC;
