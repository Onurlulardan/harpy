-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_demo_readback_v1(p_receipt_id uuid)
RETURNS TABLE(
    receipt_id uuid, package_id uuid, package_slug text,
    package_version text, release_id uuid, component_id uuid,
    owner_generation bigint, contract text, content_sha256 text,
    module_slug text, source_contract text, semantic_sha256 bytea,
    form_count integer, row_count integer, document_bytes bytea,
    idempotency_key text, request_sha256 bytea, lifecycle_state text,
    ready boolean, retry_at timestamptz, completed_at timestamptz,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN QUERY
    SELECT provision.receipt_id, provision.package_id,
        provision.package_slug, provision.package_version,
        provision.release_id, provision.component_id,
        provision.owner_generation, provision.contract,
        provision.content_sha256, payload.module_slug,
        payload.source_contract, payload.semantic_sha256,
        payload.form_count, payload.row_count, payload.document_bytes,
        provision.idempotency_key, provision.request_sha256,
        CASE
          WHEN component.lifecycle_state = 'AVAILABLE'
           AND head.component_id IS NULL THEN 'AVAILABLE'
          ELSE 'STAGED'
        END,
        provision.ready AND handoff.handoff_id IS NOT NULL,
        provision.retry_at, provision.completed_at,
        provision.created_at, provision.updated_at
    FROM catalog_demo_provisions AS provision
    JOIN catalog_demo_payloads AS payload
      ON payload.package_id = provision.package_id
     AND payload.package_version = provision.package_version
     AND payload.component_id = provision.component_id
     AND payload.release_id = provision.release_id
    JOIN catalog_catalog_components AS component
      ON component.component_id = provision.component_id
    LEFT JOIN catalog_demo_component_heads AS head
      ON head.component_id = provision.component_id
    LEFT JOIN catalog_demo_handoffs AS handoff
      ON handoff.provision_receipt_id = provision.receipt_id
    WHERE provision.receipt_id = p_receipt_id
      AND component.lifecycle_state IN (
          'AVAILABLE','STAGED','ACTIVE','DISABLED'
      );
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION catalog_demo_readback_v1(uuid) FROM PUBLIC;
