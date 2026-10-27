-- +goose Up
CREATE TABLE analytics_catalog_version_projections (
    receipt_id uuid PRIMARY KEY,
    package_id uuid NOT NULL,
    package_slug text NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    component_id uuid NOT NULL,
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    contract text NOT NULL CHECK (contract = 'analytics.catalog-definition.v1'),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    component_kind text NOT NULL CHECK (component_kind IN ('ENDPOINT','DASHBOARD')),
    analytics_kind text NOT NULL CHECK (analytics_kind IN ('endpoint','dashboard')),
    module_slug text NOT NULL,
    semantic_identity text NOT NULL,
    semantic_sha256 bytea NOT NULL CHECK (octet_length(semantic_sha256) = 32),
    item_count integer NOT NULL CHECK (item_count = 1),
    auxiliary_count integer NOT NULL CHECK (auxiliary_count > 1),
    name text NOT NULL CHECK (octet_length(name) BETWEEN 1 AND 100),
    description text NOT NULL CHECK (octet_length(description) BETWEEN 1 AND 2048),
    legacy_hash text NOT NULL CHECK (legacy_hash ~ '^[0-9a-f]{64}$'),
    definition_json jsonb NOT NULL CHECK (jsonb_typeof(definition_json) = 'object'),
    definition_sha256 bytea NOT NULL CHECK (octet_length(definition_sha256) = 32),
    document_bytes bytea NOT NULL,
    runtime_plan_bytes bytea NOT NULL,
    runtime_plan_sha256 bytea NOT NULL CHECK (octet_length(runtime_plan_sha256) = 32),
    operational_state text NOT NULL CHECK (operational_state IN ('pending','ready')),
    blocker_code text,
    operational_resource_id uuid,
    published_version_id uuid,
    current_draft_version_id uuid,
    ready_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (package_id, package_version, component_id),
    UNIQUE (component_id, release_id),
    UNIQUE (
        receipt_id, component_id, package_id, package_version,
        release_id, content_sha256
    ),
    FOREIGN KEY (component_id)
        REFERENCES analytics_catalog_components(component_id) ON DELETE RESTRICT,
    CHECK (package_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (package_slug = module_slug),
    CHECK (package_slug ~ '^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$'),
    CHECK (package_version ~
        '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$'),
    CHECK (pg_catalog.uuid_extract_version(receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (semantic_identity ~ '^[a-z][a-z0-9-]{1,62}$'),
    CHECK (date_trunc('microseconds', created_at) = created_at),
    CHECK (date_trunc('microseconds', updated_at) = updated_at),
    CHECK (updated_at >= created_at),
    CHECK (ready_at IS NULL OR (
        date_trunc('microseconds', ready_at) = ready_at AND
        ready_at >= created_at
    )),
    CHECK (
        (operational_state = 'pending' AND blocker_code ~ '^[A-Z][A-Z0-9_]{0,127}$'
         AND operational_resource_id IS NULL AND published_version_id IS NULL
         AND current_draft_version_id IS NULL AND ready_at IS NULL)
        OR
        (operational_state = 'ready' AND blocker_code IS NULL
         AND pg_catalog.uuid_extract_version(operational_resource_id) = 7
         AND pg_catalog.uuid_extract_version(published_version_id) = 7
         AND pg_catalog.uuid_extract_version(current_draft_version_id) = 7
        AND ready_at IS NOT NULL)
    )
);

CREATE TABLE analytics_catalog_projection_attempts (
    receipt_id uuid PRIMARY KEY,
    projection_receipt_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    request_sha256 bytea NOT NULL CHECK (octet_length(request_sha256) = 32),
    ready_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (receipt_id,projection_receipt_id),
    UNIQUE (projection_receipt_id,idempotency_key),
    FOREIGN KEY (projection_receipt_id)
        REFERENCES analytics_catalog_version_projections(receipt_id)
        ON DELETE RESTRICT,
    CHECK (pg_catalog.uuid_extract_version(receipt_id) = 7),
    CHECK (actor_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (
        length(idempotency_key) BETWEEN 1 AND 128 AND
        idempotency_key ~ '^[A-Za-z0-9_.:-]+$'
    ),
    CHECK (date_trunc('microseconds',created_at) = created_at),
    CHECK (date_trunc('microseconds',updated_at) = updated_at),
    CHECK (updated_at >= created_at),
    CHECK (ready_at IS NULL OR (
        date_trunc('microseconds',ready_at) = ready_at
        AND ready_at >= created_at
    ))
);
