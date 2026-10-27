-- +goose Up
INSERT INTO analytics_catalog_version_projections(
    receipt_id,package_id,package_slug,package_version,release_id,
    component_id,owner_generation,contract,content_sha256,
    component_kind,analytics_kind,module_slug,semantic_identity,
    semantic_sha256,item_count,auxiliary_count,name,description,
    legacy_hash,definition_json,definition_sha256,document_bytes,
    runtime_plan_bytes,runtime_plan_sha256,operational_state,blocker_code,
    operational_resource_id,published_version_id,current_draft_version_id,
    ready_at,created_at,updated_at
)
SELECT
    provision.receipt_id,provision.package_id,provision.package_slug,
    provision.package_version,provision.release_id,provision.component_id,
    provision.owner_generation,provision.contract,provision.content_sha256,
    provision.component_kind,definition.analytics_kind,
    provision.module_slug,provision.semantic_identity,
    provision.semantic_sha256,provision.item_count,provision.auxiliary_count,
    definition.name,definition.description,definition.legacy_hash,
    definition.definition_json,definition.definition_sha256,
    definition.document_bytes,definition.runtime_plan_bytes,
    definition.runtime_plan_sha256,definition.operational_state,
    definition.blocker_code,definition.operational_resource_id,
    definition.published_version_id,definition.current_draft_version_id,
    provision.ready_at,provision.created_at,provision.updated_at
FROM analytics_catalog_provisions AS provision
JOIN analytics_catalog_definitions AS definition
  ON definition.component_id=provision.component_id;

INSERT INTO analytics_catalog_projection_attempts(
    receipt_id,projection_receipt_id,actor_id,idempotency_key,
    request_sha256,ready_at,created_at,updated_at
)
SELECT receipt_id,receipt_id,actor_id,idempotency_key,request_sha256,
    ready_at,created_at,updated_at
FROM analytics_catalog_provisions;

CREATE TABLE analytics_catalog_component_heads (
    component_id uuid PRIMARY KEY,
    projection_receipt_id uuid NOT NULL UNIQUE,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    activated_at timestamptz NOT NULL,
    FOREIGN KEY (
        projection_receipt_id,component_id,package_id,package_version,
        release_id,content_sha256
    ) REFERENCES analytics_catalog_version_projections(
        receipt_id,component_id,package_id,package_version,
        release_id,content_sha256
    ) ON DELETE RESTRICT,
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(projection_receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (date_trunc('microseconds', activated_at) = activated_at)
);

INSERT INTO analytics_catalog_component_heads(
    component_id,projection_receipt_id,package_id,package_version,
    release_id,content_sha256,activated_at
)
SELECT
    projection.component_id,projection.receipt_id,projection.package_id,
    projection.package_version,projection.release_id,
    projection.content_sha256,
    GREATEST(projection.ready_at,component.updated_at)
FROM analytics_catalog_version_projections AS projection
JOIN analytics_catalog_components AS component
  ON component.component_id=projection.component_id
WHERE projection.ready_at IS NOT NULL
  AND component.lifecycle_state IN ('ACTIVE','DISABLED','RETIRED','PURGED');
