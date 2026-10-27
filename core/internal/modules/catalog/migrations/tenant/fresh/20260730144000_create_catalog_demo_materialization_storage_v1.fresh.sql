-- +goose Up
CREATE TABLE catalog_demo_payloads (
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    package_slug text NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    module_slug text NOT NULL,
    source_contract text NOT NULL,
    semantic_sha256 bytea NOT NULL CHECK (octet_length(semantic_sha256) = 32),
    form_count integer NOT NULL CHECK (form_count > 0),
    row_count integer NOT NULL CHECK (row_count > 0),
    document_bytes bytea NOT NULL CHECK (octet_length(document_bytes) > 0),
    installed_at timestamptz NOT NULL,
    PRIMARY KEY (package_id, package_version, component_id),
    UNIQUE (component_id, package_version),
    UNIQUE (component_id, release_id),
    UNIQUE (
        component_id, package_id, package_version, release_id,
        form_count, row_count
    ),
    FOREIGN KEY (component_id)
        REFERENCES catalog_catalog_components(component_id) ON DELETE RESTRICT,
    CHECK (package_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (package_slug = module_slug),
    CHECK (package_slug ~ '^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$'),
    CHECK (package_version ~
        '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$'),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (date_trunc('microseconds', installed_at) = installed_at)
);

CREATE TABLE catalog_demo_provisions (
    receipt_id uuid PRIMARY KEY,
    package_id uuid NOT NULL,
    package_slug text NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    component_id uuid NOT NULL,
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    contract text NOT NULL CHECK (contract = 'catalog.demo-data.v1'),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 128),
    request_sha256 bytea NOT NULL CHECK (octet_length(request_sha256) = 32),
    ready boolean NOT NULL,
    retry_at timestamptz,
    completed_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (package_id, package_version, component_id),
    UNIQUE (
        receipt_id, component_id, package_id, package_version,
        release_id, content_sha256
    ),
    UNIQUE (
        receipt_id, component_id, package_id, package_version,
        release_id, content_sha256, request_sha256
    ),
    FOREIGN KEY (package_id, package_version, component_id)
        REFERENCES catalog_demo_payloads(
            package_id, package_version, component_id
        ) ON DELETE RESTRICT,
    CHECK (package_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (package_slug ~ '^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$'),
    CHECK (pg_catalog.uuid_extract_version(receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (idempotency_key ~ '^[A-Za-z0-9_.:-]+$'),
    CHECK (ready = (retry_at IS NULL)),
    CHECK (retry_at IS NULL OR retry_at > completed_at),
    CHECK (completed_at >= created_at AND updated_at >= created_at),
    CHECK (date_trunc('microseconds', completed_at) = completed_at),
    CHECK (date_trunc('microseconds', created_at) = created_at),
    CHECK (date_trunc('microseconds', updated_at) = updated_at),
    CHECK (retry_at IS NULL OR date_trunc('microseconds', retry_at) = retry_at)
);

CREATE TABLE catalog_demo_handoffs (
    handoff_id uuid PRIMARY KEY,
    provision_receipt_id uuid NOT NULL UNIQUE,
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    forms_run_id uuid NOT NULL,
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    request_sha256 bytea NOT NULL CHECK (octet_length(request_sha256) = 32),
    form_count integer NOT NULL CHECK (form_count > 0),
    row_count integer NOT NULL CHECK (row_count > 0),
    readback_sha256 bytea NOT NULL CHECK (octet_length(readback_sha256) = 32),
    completed_at timestamptz NOT NULL,
    UNIQUE (provision_receipt_id, forms_run_id),
    UNIQUE (
        handoff_id, provision_receipt_id, component_id, package_id,
        package_version, release_id, forms_run_id, content_sha256
    ),
    FOREIGN KEY (
        provision_receipt_id, component_id, package_id, package_version,
        release_id, content_sha256, request_sha256
    ) REFERENCES catalog_demo_provisions (
        receipt_id, component_id, package_id, package_version,
        release_id, content_sha256, request_sha256
    ) ON DELETE RESTRICT,
    FOREIGN KEY (
        component_id, package_id, package_version, release_id,
        form_count, row_count
    )
        REFERENCES catalog_demo_payloads (
            component_id, package_id, package_version, release_id,
            form_count, row_count
        ) ON DELETE RESTRICT,
    CHECK (pg_catalog.uuid_extract_version(handoff_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(provision_receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(forms_run_id) = 7),
    CHECK (readback_sha256 <> decode(repeat('00', 32), 'hex')),
    CHECK (date_trunc('microseconds', completed_at) = completed_at)
);

CREATE TABLE catalog_demo_component_heads (
    component_id uuid PRIMARY KEY,
    package_id uuid NOT NULL,
    package_version text NOT NULL,
    release_id uuid NOT NULL,
    provision_receipt_id uuid NOT NULL UNIQUE,
    handoff_id uuid NOT NULL UNIQUE,
    forms_run_id uuid NOT NULL,
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    activated_at timestamptz NOT NULL,
    FOREIGN KEY (
        provision_receipt_id, component_id, package_id, package_version,
        release_id, content_sha256
    ) REFERENCES catalog_demo_provisions (
        receipt_id, component_id, package_id, package_version,
        release_id, content_sha256
    ) ON DELETE RESTRICT,
    FOREIGN KEY (
        handoff_id, provision_receipt_id, component_id, package_id,
        package_version, release_id, forms_run_id, content_sha256
    ) REFERENCES catalog_demo_handoffs (
        handoff_id, provision_receipt_id, component_id, package_id,
        package_version, release_id, forms_run_id, content_sha256
    ) ON DELETE RESTRICT,
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(forms_run_id) = 7),
    CHECK (date_trunc('microseconds', activated_at) = activated_at)
);
