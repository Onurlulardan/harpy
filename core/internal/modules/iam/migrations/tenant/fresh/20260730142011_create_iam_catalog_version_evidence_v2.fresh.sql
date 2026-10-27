-- +goose Up
CREATE TABLE iam_catalog_version_evidence (
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    package_slug text NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    contract text NOT NULL CHECK (contract = 'iam.catalog-template.v1'),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    component_kind text NOT NULL CHECK (component_kind = 'IAM_TEMPLATE'),
    module_slug text NOT NULL,
    semantic_identity text NOT NULL,
    semantic_sha256 bytea NOT NULL CHECK (octet_length(semantic_sha256) = 32),
    item_count integer NOT NULL CHECK (item_count > 1),
    auxiliary_count integer NOT NULL CHECK (auxiliary_count > 0),
    document_bytes bytea NOT NULL CHECK (octet_length(document_bytes) > 0),
    installed_at timestamptz NOT NULL,
    PRIMARY KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256
    ),
    UNIQUE (package_id, package_version, component_id),
    UNIQUE (component_id, release_id),
    FOREIGN KEY (component_id)
        REFERENCES iam_catalog_components(component_id) ON DELETE RESTRICT,
    CHECK (package_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (package_slug = module_slug AND semantic_identity = module_slug),
    CHECK (package_slug ~ '^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$'),
    CHECK (package_version ~
        '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$'),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (date_trunc('microseconds', installed_at) = installed_at)
);

CREATE TABLE iam_catalog_version_resources (
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    owner_generation bigint NOT NULL,
    content_sha256 text NOT NULL,
    resource_slug text NOT NULL,
    resource_kind text NOT NULL CHECK (resource_kind IN ('module','form','workflow')),
    name text NOT NULL,
    description text NOT NULL,
    PRIMARY KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug
    ),
    FOREIGN KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256
    ) REFERENCES iam_catalog_version_evidence (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256
    ) ON DELETE RESTRICT,
    CHECK (resource_slug ~ '^(module|form|workflow)\.[a-z][a-z0-9-]{1,62}$'),
    CHECK (resource_kind = split_part(resource_slug, '.', 1))
);

CREATE TABLE iam_catalog_version_grants (
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    owner_generation bigint NOT NULL,
    content_sha256 text NOT NULL,
    resource_slug text NOT NULL,
    role_name text NOT NULL,
    role_description text NOT NULL,
    actions text[] NOT NULL CHECK (cardinality(actions) > 0),
    PRIMARY KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug, role_name
    ),
    FOREIGN KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug
    ) REFERENCES iam_catalog_version_resources (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug
    ) ON DELETE RESTRICT,
    CHECK (role_name <> '' AND role_description <> '')
);

CREATE TABLE iam_catalog_version_provisions (
    receipt_id uuid PRIMARY KEY,
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    package_slug text NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    owner_generation bigint NOT NULL,
    contract text NOT NULL,
    content_sha256 text NOT NULL,
    component_kind text NOT NULL,
    module_slug text NOT NULL,
    semantic_identity text NOT NULL,
    semantic_sha256 bytea NOT NULL CHECK (octet_length(semantic_sha256) = 32),
    item_count integer NOT NULL,
    auxiliary_count integer NOT NULL,
    idempotency_key text NOT NULL,
    request_sha256 bytea NOT NULL CHECK (octet_length(request_sha256) = 32),
    ready_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (component_id, idempotency_key),
    UNIQUE (
        receipt_id, component_id, package_id, package_version, release_id,
        owner_generation, content_sha256
    ),
    FOREIGN KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256
    ) REFERENCES iam_catalog_version_evidence (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256
    ) ON DELETE RESTRICT,
    CHECK (pg_catalog.uuid_extract_version(receipt_id) = 7),
    CHECK (length(idempotency_key) BETWEEN 1 AND 128),
    CHECK (idempotency_key ~ '^[A-Za-z0-9_.:-]+$'),
    CHECK (ready_at >= created_at AND updated_at >= created_at),
    CHECK (date_trunc('microseconds', ready_at) = ready_at),
    CHECK (date_trunc('microseconds', created_at) = created_at),
    CHECK (date_trunc('microseconds', updated_at) = updated_at)
);
