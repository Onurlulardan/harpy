-- +goose Up
CREATE TABLE iam_catalog_version_operational_resources (
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    owner_generation bigint NOT NULL,
    content_sha256 text NOT NULL,
    resource_slug text NOT NULL,
    resource_id uuid NOT NULL,
    PRIMARY KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug
    ),
    UNIQUE (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug, resource_id
    ),
    FOREIGN KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug
    ) REFERENCES iam_catalog_version_resources (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug
    ) ON DELETE RESTRICT,
    FOREIGN KEY (resource_id) REFERENCES iam_resources(id) ON DELETE RESTRICT
);

CREATE TABLE iam_catalog_version_operational_grants (
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    owner_generation bigint NOT NULL,
    content_sha256 text NOT NULL,
    resource_slug text NOT NULL,
    role_name text NOT NULL,
    action_key text NOT NULL,
    role_id uuid NOT NULL,
    resource_id uuid NOT NULL,
    action_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    binding_id uuid NOT NULL,
    PRIMARY KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug, role_name, action_key
    ),
    FOREIGN KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug, role_name
    ) REFERENCES iam_catalog_version_grants (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug, role_name
    ) ON DELETE RESTRICT,
    FOREIGN KEY (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug, resource_id
    ) REFERENCES iam_catalog_version_operational_resources (
        component_id, package_id, package_version, release_id,
        owner_generation, content_sha256, resource_slug, resource_id
    ) ON DELETE RESTRICT,
    FOREIGN KEY (role_id) REFERENCES iam_roles(id) ON DELETE RESTRICT,
    FOREIGN KEY (action_id) REFERENCES iam_actions(id) ON DELETE RESTRICT,
    FOREIGN KEY (permission_id) REFERENCES iam_permissions(id) ON DELETE RESTRICT,
    FOREIGN KEY (binding_id)
        REFERENCES iam_role_permission_bindings(id) ON DELETE RESTRICT
);

CREATE TABLE iam_catalog_component_candidates (
    component_id uuid PRIMARY KEY,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    owner_generation bigint NOT NULL,
    content_sha256 text NOT NULL,
    provision_receipt_id uuid NOT NULL UNIQUE,
    staged_operation_id uuid,
    staged_at timestamptz NOT NULL,
    FOREIGN KEY (
        provision_receipt_id, component_id, package_id, package_version,
        release_id, owner_generation, content_sha256
    ) REFERENCES iam_catalog_version_provisions (
        receipt_id, component_id, package_id, package_version,
        release_id, owner_generation, content_sha256
    ) ON DELETE RESTRICT,
    CHECK (staged_operation_id IS NULL OR
        pg_catalog.uuid_extract_version(staged_operation_id) = 7),
    CHECK (date_trunc('microseconds', staged_at) = staged_at)
);

CREATE TABLE iam_catalog_component_heads (
    component_id uuid PRIMARY KEY,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    owner_generation bigint NOT NULL,
    content_sha256 text NOT NULL,
    provision_receipt_id uuid NOT NULL,
    activated_at timestamptz NOT NULL,
    FOREIGN KEY (
        provision_receipt_id, component_id, package_id, package_version,
        release_id, owner_generation, content_sha256
    ) REFERENCES iam_catalog_version_provisions (
        receipt_id, component_id, package_id, package_version,
        release_id, owner_generation, content_sha256
    ) ON DELETE RESTRICT,
    CHECK (date_trunc('microseconds', activated_at) = activated_at)
);
