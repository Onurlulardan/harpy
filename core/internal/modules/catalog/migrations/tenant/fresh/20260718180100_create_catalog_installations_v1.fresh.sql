-- +goose Up
CREATE TABLE catalog_installations (
    installation_id uuid PRIMARY KEY,
    package_id uuid NOT NULL UNIQUE,
    active_version text,
    state text NOT NULL CHECK (state IN (
        'INSTALLING','ACTIVE','DISABLED','INSTALL_FAILED','RETIRED','PURGE_PENDING','PURGED'
    )),
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 0),
    row_version bigint NOT NULL CHECK (row_version > 0),
    updated_at timestamptz NOT NULL,
    CHECK ((state IN ('INSTALLING','INSTALL_FAILED') AND active_version IS NULL)
        OR (state NOT IN ('INSTALLING','INSTALL_FAILED') AND active_version IS NOT NULL)),
    CHECK (date_trunc('microseconds', updated_at) = updated_at)
);

CREATE TABLE catalog_component_bindings (
    installation_id uuid NOT NULL,
    component_id uuid NOT NULL,
    owner_module text NOT NULL CHECK (owner_module IN ('iam','forms','workflow','analytics','catalog')),
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    owner_contract text NOT NULL CHECK (length(owner_contract) BETWEEN 1 AND 64),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    state text NOT NULL CHECK (state IN ('STAGED','ACTIVE','DISABLED','RETIRED','PURGED')),
    PRIMARY KEY (installation_id, component_id),
    CONSTRAINT catalog_component_bindings_installation_fk
        FOREIGN KEY (installation_id) REFERENCES catalog_installations(installation_id)
        ON DELETE RESTRICT
);

CREATE TABLE catalog_dependency_resolutions (
    installation_id uuid NOT NULL,
    dependency_package_id uuid NOT NULL,
    exact_version text NOT NULL,
    required boolean NOT NULL,
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    PRIMARY KEY (installation_id, dependency_package_id),
    CONSTRAINT catalog_dependency_resolutions_installation_fk
        FOREIGN KEY (installation_id) REFERENCES catalog_installations(installation_id)
        ON DELETE RESTRICT
);

CREATE INDEX catalog_dependency_resolutions_reverse_idx
ON catalog_dependency_resolutions (dependency_package_id, installation_id);
