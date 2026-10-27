-- +goose Up
CREATE TABLE iam_catalog_materializations (
    component_id uuid PRIMARY KEY,
    module_slug text NOT NULL UNIQUE,
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    semantic_sha256 bytea NOT NULL CHECK (octet_length(semantic_sha256) = 32),
    document_bytes bytea NOT NULL,
    package_id uuid NOT NULL,
    package_slug text NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    installed_at timestamptz NOT NULL,
    FOREIGN KEY (component_id)
        REFERENCES iam_catalog_components(component_id) ON DELETE RESTRICT,
    CHECK (module_slug = package_slug),
    CHECK (module_slug ~ '^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$'),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (date_trunc('microseconds', installed_at) = installed_at)
);

CREATE TABLE iam_catalog_resource_projections (
    component_id uuid NOT NULL,
    resource_slug text NOT NULL UNIQUE,
    resource_kind text NOT NULL CHECK (resource_kind IN ('module','form','workflow')),
    name text NOT NULL,
    description text NOT NULL,
    PRIMARY KEY (component_id, resource_slug),
    FOREIGN KEY (component_id)
        REFERENCES iam_catalog_materializations(component_id) ON DELETE RESTRICT,
    CHECK (resource_slug ~ '^(module|form|workflow)\.[a-z][a-z0-9-]{1,62}$'),
    CHECK (resource_kind = split_part(resource_slug, '.', 1))
);

CREATE TABLE iam_catalog_grant_projections (
    component_id uuid NOT NULL,
    resource_slug text NOT NULL,
    role_name text NOT NULL,
    role_description text NOT NULL,
    actions text[] NOT NULL CHECK (cardinality(actions) > 0),
    PRIMARY KEY (component_id, resource_slug, role_name),
    UNIQUE (resource_slug, role_name),
    FOREIGN KEY (component_id, resource_slug)
        REFERENCES iam_catalog_resource_projections(component_id, resource_slug)
        ON DELETE RESTRICT,
    CHECK (role_name <> ''),
    CHECK (role_description <> '')
);

CREATE TABLE iam_catalog_operational_resources (
    component_id uuid NOT NULL,
    resource_slug text NOT NULL,
    resource_id uuid NOT NULL,
    PRIMARY KEY (component_id, resource_slug),
    UNIQUE (resource_id),
    FOREIGN KEY (component_id, resource_slug)
        REFERENCES iam_catalog_resource_projections(component_id, resource_slug)
        ON DELETE RESTRICT,
    FOREIGN KEY (resource_id) REFERENCES iam_resources(id) ON DELETE RESTRICT
);

CREATE TABLE iam_catalog_operational_grants (
    component_id uuid NOT NULL,
    resource_slug text NOT NULL,
    role_name text NOT NULL,
    action_key text NOT NULL,
    role_id uuid NOT NULL,
    resource_id uuid NOT NULL,
    action_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    binding_id uuid NOT NULL,
    PRIMARY KEY (component_id, resource_slug, role_name, action_key),
    UNIQUE (binding_id),
    FOREIGN KEY (component_id, resource_slug, role_name)
        REFERENCES iam_catalog_grant_projections(
            component_id, resource_slug, role_name
        ) ON DELETE RESTRICT,
    FOREIGN KEY (role_id) REFERENCES iam_roles(id) ON DELETE RESTRICT,
    FOREIGN KEY (resource_id) REFERENCES iam_resources(id) ON DELETE RESTRICT,
    FOREIGN KEY (action_id) REFERENCES iam_actions(id) ON DELETE RESTRICT,
    FOREIGN KEY (permission_id) REFERENCES iam_permissions(id) ON DELETE RESTRICT,
    FOREIGN KEY (binding_id)
        REFERENCES iam_role_permission_bindings(id) ON DELETE RESTRICT
);

CREATE TABLE iam_catalog_provisions (
    receipt_id uuid PRIMARY KEY,
    package_id uuid NOT NULL,
    package_slug text NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    component_id uuid NOT NULL,
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    contract text NOT NULL CHECK (contract = 'iam.catalog-template.v1'),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    component_kind text NOT NULL CHECK (component_kind = 'IAM_TEMPLATE'),
    module_slug text NOT NULL,
    semantic_identity text NOT NULL,
    semantic_sha256 bytea NOT NULL CHECK (octet_length(semantic_sha256) = 32),
    item_count integer NOT NULL CHECK (item_count > 1),
    auxiliary_count integer NOT NULL CHECK (auxiliary_count > 0),
    idempotency_key text NOT NULL,
    request_sha256 bytea NOT NULL CHECK (octet_length(request_sha256) = 32),
    ready_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (component_id, idempotency_key),
    FOREIGN KEY (component_id)
        REFERENCES iam_catalog_materializations(component_id) ON DELETE RESTRICT,
    CHECK (module_slug = package_slug),
    CHECK (semantic_identity = module_slug),
    CHECK (package_slug ~ '^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$'),
    CHECK (pg_catalog.uuid_extract_version(receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (length(idempotency_key) BETWEEN 1 AND 128),
    CHECK (idempotency_key ~ '^[A-Za-z0-9_.:-]+$'),
    CHECK (date_trunc('microseconds', ready_at) = ready_at),
    CHECK (date_trunc('microseconds', created_at) = created_at),
    CHECK (date_trunc('microseconds', updated_at) = updated_at),
    CHECK (ready_at >= created_at AND updated_at >= created_at)
);
