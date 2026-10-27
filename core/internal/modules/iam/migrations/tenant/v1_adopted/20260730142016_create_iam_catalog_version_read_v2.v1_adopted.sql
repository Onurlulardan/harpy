-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_projection_read_v2(p_receipt_id uuid)
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
    SELECT provision.receipt_id,provision.package_id,
        provision.package_slug,provision.package_version,
        provision.release_id,provision.component_id,
        provision.owner_generation,provision.contract,
        provision.content_sha256,provision.component_kind,
        provision.module_slug,provision.semantic_identity,
        provision.semantic_sha256,provision.item_count,
        provision.auxiliary_count,evidence.document_bytes,
        provision.idempotency_key,provision.request_sha256,
        CASE
          WHEN candidate.provision_receipt_id=provision.receipt_id
           AND head.component_id IS NULL
           AND component.lifecycle_state='AVAILABLE' THEN 'AVAILABLE'
          ELSE 'STAGED'
        END,
        provision.ready_at,provision.created_at,provision.updated_at
    FROM iam_catalog_version_provisions AS provision
    JOIN iam_catalog_version_evidence AS evidence
      ON evidence.component_id=provision.component_id
     AND evidence.package_id=provision.package_id
     AND evidence.package_version=provision.package_version
     AND evidence.release_id=provision.release_id
     AND evidence.owner_generation=provision.owner_generation
     AND evidence.content_sha256=provision.content_sha256
    JOIN iam_catalog_components AS component
      ON component.component_id=provision.component_id
    LEFT JOIN iam_catalog_component_candidates AS candidate
      ON candidate.component_id=provision.component_id
    LEFT JOIN iam_catalog_component_heads AS head
      ON head.component_id=provision.component_id
    WHERE provision.receipt_id=p_receipt_id
      AND component.lifecycle_state<>'PURGED';
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_projection_read_v2(uuid) FROM PUBLIC;
