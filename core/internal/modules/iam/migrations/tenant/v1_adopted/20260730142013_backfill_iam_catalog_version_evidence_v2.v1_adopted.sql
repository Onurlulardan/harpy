-- +goose Up
INSERT INTO iam_catalog_version_evidence(
    component_id,package_id,package_slug,package_version,release_id,
    owner_generation,contract,content_sha256,component_kind,module_slug,
    semantic_identity,semantic_sha256,item_count,auxiliary_count,
    document_bytes,installed_at
)
SELECT DISTINCT
    provision.component_id,provision.package_id,provision.package_slug,
    provision.package_version,provision.release_id,
    provision.owner_generation,provision.contract,
    provision.content_sha256,provision.component_kind,
    provision.module_slug,provision.semantic_identity,
    provision.semantic_sha256,provision.item_count,
    provision.auxiliary_count,materialization.document_bytes,
    materialization.installed_at
FROM iam_catalog_provisions AS provision
JOIN iam_catalog_materializations AS materialization
  ON materialization.component_id=provision.component_id;

INSERT INTO iam_catalog_version_resources(
    component_id,package_id,package_version,release_id,
    owner_generation,content_sha256,resource_slug,
    resource_kind,name,description
)
SELECT evidence.component_id,evidence.package_id,evidence.package_version,
    evidence.release_id,evidence.owner_generation,evidence.content_sha256,
    resource.resource_slug,resource.resource_kind,
    resource.name,resource.description
FROM iam_catalog_version_evidence AS evidence
JOIN iam_catalog_resource_projections AS resource
  ON resource.component_id=evidence.component_id;

INSERT INTO iam_catalog_version_grants(
    component_id,package_id,package_version,release_id,
    owner_generation,content_sha256,resource_slug,
    role_name,role_description,actions
)
SELECT evidence.component_id,evidence.package_id,evidence.package_version,
    evidence.release_id,evidence.owner_generation,evidence.content_sha256,
    grant_row.resource_slug,grant_row.role_name,
    grant_row.role_description,grant_row.actions
FROM iam_catalog_version_evidence AS evidence
JOIN iam_catalog_grant_projections AS grant_row
  ON grant_row.component_id=evidence.component_id;

INSERT INTO iam_catalog_version_provisions(
    receipt_id,component_id,package_id,package_slug,package_version,
    release_id,owner_generation,contract,content_sha256,component_kind,
    module_slug,semantic_identity,semantic_sha256,item_count,
    auxiliary_count,idempotency_key,request_sha256,
    ready_at,created_at,updated_at
)
SELECT provision.receipt_id,provision.component_id,provision.package_id,
    provision.package_slug,provision.package_version,provision.release_id,
    provision.owner_generation,provision.contract,
    provision.content_sha256,provision.component_kind,
    provision.module_slug,provision.semantic_identity,
    provision.semantic_sha256,provision.item_count,
    provision.auxiliary_count,provision.idempotency_key,
    provision.request_sha256,provision.ready_at,
    provision.created_at,provision.updated_at
FROM iam_catalog_provisions AS provision;

INSERT INTO iam_catalog_version_operational_resources(
    component_id,package_id,package_version,release_id,
    owner_generation,content_sha256,resource_slug,resource_id
)
SELECT evidence.component_id,evidence.package_id,evidence.package_version,
    evidence.release_id,evidence.owner_generation,evidence.content_sha256,
    resource.resource_slug,resource.resource_id
FROM iam_catalog_version_evidence AS evidence
JOIN iam_catalog_operational_resources AS resource
  ON resource.component_id=evidence.component_id;

INSERT INTO iam_catalog_version_operational_grants(
    component_id,package_id,package_version,release_id,
    owner_generation,content_sha256,resource_slug,role_name,action_key,
    role_id,resource_id,action_id,permission_id,binding_id
)
SELECT evidence.component_id,evidence.package_id,evidence.package_version,
    evidence.release_id,evidence.owner_generation,evidence.content_sha256,
    grant_row.resource_slug,grant_row.role_name,grant_row.action_key,
    grant_row.role_id,grant_row.resource_id,grant_row.action_id,
    grant_row.permission_id,grant_row.binding_id
FROM iam_catalog_version_evidence AS evidence
JOIN iam_catalog_operational_grants AS grant_row
  ON grant_row.component_id=evidence.component_id;

INSERT INTO iam_catalog_component_heads(
    component_id,package_id,package_version,release_id,
    owner_generation,content_sha256,provision_receipt_id,activated_at
)
SELECT component.component_id,provision.package_id,
    provision.package_version,provision.release_id,
    provision.owner_generation,provision.content_sha256,
    provision.receipt_id,provision.ready_at
FROM iam_catalog_components AS component
JOIN LATERAL (
    SELECT candidate.*
    FROM iam_catalog_version_provisions AS candidate
    WHERE candidate.component_id=component.component_id
      AND candidate.content_sha256=component.content_sha256
    ORDER BY candidate.created_at DESC,candidate.receipt_id DESC
    LIMIT 1
) AS provision ON true
WHERE component.lifecycle_state IN ('ACTIVE','DISABLED','RETIRED','PURGED');

INSERT INTO iam_catalog_component_candidates(
    component_id,package_id,package_version,release_id,
    owner_generation,content_sha256,provision_receipt_id,staged_at
)
SELECT component.component_id,provision.package_id,
    provision.package_version,provision.release_id,
    provision.owner_generation,provision.content_sha256,
    provision.receipt_id,provision.ready_at
FROM iam_catalog_components AS component
JOIN LATERAL (
    SELECT candidate.*
    FROM iam_catalog_version_provisions AS candidate
    WHERE candidate.component_id=component.component_id
      AND candidate.content_sha256=component.content_sha256
    ORDER BY candidate.created_at DESC,candidate.receipt_id DESC
    LIMIT 1
) AS provision ON true
WHERE component.lifecycle_state IN ('AVAILABLE','STAGED');
