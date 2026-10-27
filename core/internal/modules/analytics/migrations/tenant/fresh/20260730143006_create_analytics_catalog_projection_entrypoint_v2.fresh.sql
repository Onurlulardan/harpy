-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_project_v1(
    p_scope_schema name, p_receipt_id uuid,
    p_package_id uuid, p_package_slug text, p_package_version text,
    p_release_id uuid, p_component_id uuid, p_owner_generation bigint,
    p_contract text, p_content_sha256 text, p_component_kind text,
    p_module_slug text, p_semantic_identity text, p_semantic_sha256 bytea,
    p_item_count integer, p_auxiliary_count integer, p_document_bytes bytea,
    p_runtime_plan_bytes bytea, p_runtime_plan_sha256 bytea, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_occurred_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid, package_id uuid, package_slug text,
    package_version text, release_id uuid,
    component_id uuid, owner_generation bigint, contract text,
    content_sha256 text, component_kind text, module_slug text,
    semantic_identity text, semantic_sha256 bytea,
    item_count integer, auxiliary_count integer, document_bytes bytea,
    idempotency_key text, request_sha256 bytea, lifecycle_state text,
    ready_at timestamptz, created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_receipt_id uuid;
BEGIN
    v_receipt_id := analytics_catalog_projection_begin_v1(
        p_scope_schema,p_receipt_id,p_package_id,p_package_slug,
        p_package_version,p_release_id,p_component_id,p_owner_generation,
        p_contract,p_content_sha256,p_component_kind,p_module_slug,
        p_semantic_identity,p_semantic_sha256,p_item_count,p_auxiliary_count,
        p_document_bytes,p_runtime_plan_bytes,p_runtime_plan_sha256,p_actor_id,
        p_idempotency_key,p_request_sha256,p_occurred_at
    );
    PERFORM analytics_catalog_projection_materialize_v1(
        p_scope_schema,v_receipt_id,p_occurred_at
    );
    RETURN QUERY
    SELECT provision.receipt_id,provision.package_id,provision.package_slug,
        provision.package_version,provision.release_id,
        provision.component_id,provision.owner_generation,provision.contract,
        provision.content_sha256,provision.component_kind,
        provision.module_slug,provision.semantic_identity,
        provision.semantic_sha256,provision.item_count,
        provision.auxiliary_count,definition.document_bytes,
        provision.idempotency_key,provision.request_sha256,
        component.lifecycle_state,provision.ready_at,
        provision.created_at,provision.updated_at
    FROM analytics_catalog_provisions AS provision
    JOIN analytics_catalog_definitions AS definition
      ON definition.component_id=provision.component_id
    JOIN analytics_catalog_components AS component
      ON component.component_id=provision.component_id
    WHERE provision.receipt_id=v_receipt_id
      AND component.lifecycle_state IN ('AVAILABLE','STAGED')
      AND (
        (provision.ready_at IS NULL AND
         definition.operational_state='pending' AND
         definition.blocker_code IS NOT NULL)
        OR
        (provision.ready_at IS NOT NULL AND
         definition.operational_state='ready' AND
         definition.blocker_code IS NULL AND
         definition.operational_resource_id IS NOT NULL AND
         definition.published_version_id IS NOT NULL AND
         definition.current_draft_version_id IS NOT NULL)
      );
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog projection readback differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) FROM PUBLIC;
