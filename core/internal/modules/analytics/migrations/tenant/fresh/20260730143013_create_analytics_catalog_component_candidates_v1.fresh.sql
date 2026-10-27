-- +goose Up
CREATE TABLE analytics_catalog_projection_activations (
    projection_receipt_id uuid PRIMARY KEY,
    attempt_receipt_id uuid NOT NULL UNIQUE,
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    activated_at timestamptz NOT NULL,
    superseded_at timestamptz,
    FOREIGN KEY (
        projection_receipt_id,component_id,package_id,package_version,
        release_id,content_sha256
    ) REFERENCES analytics_catalog_version_projections(
        receipt_id,component_id,package_id,package_version,
        release_id,content_sha256
    ) ON DELETE RESTRICT,
    FOREIGN KEY (attempt_receipt_id,projection_receipt_id)
        REFERENCES analytics_catalog_projection_attempts(
            receipt_id,projection_receipt_id
        ) ON DELETE RESTRICT,
    CHECK (pg_catalog.uuid_extract_version(projection_receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(attempt_receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (date_trunc('microseconds', activated_at) = activated_at),
    CHECK (superseded_at IS NULL OR (
        date_trunc('microseconds', superseded_at) = superseded_at
        AND superseded_at >= activated_at
    ))
);

INSERT INTO analytics_catalog_projection_activations(
    projection_receipt_id,attempt_receipt_id,
    component_id,package_id,package_version,
    release_id,content_sha256,activated_at
)
SELECT head.projection_receipt_id,attempt.receipt_id,
    head.component_id,head.package_id,
    head.package_version,head.release_id,head.content_sha256,head.activated_at
FROM analytics_catalog_component_heads AS head
JOIN analytics_catalog_projection_attempts AS attempt
  ON attempt.projection_receipt_id=head.projection_receipt_id;

CREATE UNIQUE INDEX analytics_catalog_projection_activations_live_key
    ON analytics_catalog_projection_activations(component_id)
    WHERE superseded_at IS NULL;

CREATE TABLE analytics_catalog_component_candidates (
    component_id uuid PRIMARY KEY,
    projection_receipt_id uuid NOT NULL UNIQUE,
    attempt_receipt_id uuid NOT NULL UNIQUE,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    staged_operation_id uuid,
    prepared_at timestamptz NOT NULL,
    FOREIGN KEY (
        projection_receipt_id,component_id,package_id,package_version,
        release_id,content_sha256
    ) REFERENCES analytics_catalog_version_projections(
        receipt_id,component_id,package_id,package_version,
        release_id,content_sha256
    ) ON DELETE RESTRICT,
    FOREIGN KEY (attempt_receipt_id,projection_receipt_id)
        REFERENCES analytics_catalog_projection_attempts(
            receipt_id,projection_receipt_id
        ) ON DELETE RESTRICT,
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(projection_receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(attempt_receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (staged_operation_id IS NULL OR
        pg_catalog.uuid_extract_version(staged_operation_id) = 7),
    CHECK (date_trunc('microseconds', prepared_at) = prepared_at)
);

INSERT INTO analytics_catalog_component_candidates(
    component_id,projection_receipt_id,attempt_receipt_id,
    package_id,package_version,
    release_id,owner_generation,content_sha256,staged_operation_id,prepared_at
)
SELECT
    projection.component_id,projection.receipt_id,attempt.receipt_id,
    projection.package_id,
    projection.package_version,projection.release_id,
    projection.owner_generation,projection.content_sha256,NULL::uuid,
    projection.ready_at
FROM analytics_catalog_version_projections AS projection
JOIN analytics_catalog_projection_attempts AS attempt
  ON attempt.projection_receipt_id=projection.receipt_id
JOIN analytics_catalog_components AS component
  ON component.component_id=projection.component_id
LEFT JOIN analytics_catalog_component_heads AS head
  ON head.component_id=projection.component_id
LEFT JOIN analytics_catalog_projection_activations AS activation
  ON activation.projection_receipt_id=projection.receipt_id
WHERE projection.operational_state='ready'
  AND projection.ready_at IS NOT NULL
  AND head.component_id IS NULL
  AND activation.projection_receipt_id IS NULL
  AND component.lifecycle_state IN ('AVAILABLE','STAGED');
