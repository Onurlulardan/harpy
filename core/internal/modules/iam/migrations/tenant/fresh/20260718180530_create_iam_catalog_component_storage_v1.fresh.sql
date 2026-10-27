-- +goose Up
CREATE TABLE iam_catalog_components (
    component_id uuid PRIMARY KEY,
    component_kind text NOT NULL CHECK (component_kind IN ('IAM_TEMPLATE')),
    contract text NOT NULL CHECK (length(contract) BETWEEN 1 AND 64),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    writer_state text NOT NULL CHECK (writer_state = 'v2'),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'AVAILABLE','STAGED','ACTIVE','DISABLED','RETIRED','PURGED'
    )),
    updated_at timestamptz NOT NULL,
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (date_trunc('microseconds', updated_at) = updated_at)
);

CREATE TABLE iam_catalog_component_receipts (
    receipt_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL,
    component_id uuid NOT NULL,
    owner_module text NOT NULL CHECK (owner_module = 'iam'),
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    contract text NOT NULL,
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    action text NOT NULL CHECK (action IN (
        'stage','cleanup','activate','disable','enable','retire','restore','purge'
    )),
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 128),
    ready boolean NOT NULL,
    completed_at timestamptz NOT NULL,
    UNIQUE (operation_id, component_id, action),
    UNIQUE (component_id, action, idempotency_key),
    FOREIGN KEY (component_id) REFERENCES iam_catalog_components(component_id) ON DELETE RESTRICT,
    CHECK (pg_catalog.uuid_extract_version(receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(operation_id) = 7),
    CHECK (date_trunc('microseconds', completed_at) = completed_at)
);

CREATE TABLE iam_catalog_component_blockers (
    component_id uuid NOT NULL,
    code text NOT NULL CHECK (code ~ '^[A-Z][A-Z0-9_]{0,127}$'),
    digest text NOT NULL CHECK (digest ~ '^[0-9a-f]{64}$'),
    active boolean NOT NULL,
    PRIMARY KEY (component_id, code),
    FOREIGN KEY (component_id) REFERENCES iam_catalog_components(component_id) ON DELETE RESTRICT
);

CREATE TABLE iam_catalog_blocker_receipts (
    receipt_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL,
    component_id uuid NOT NULL,
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 0),
    catalog_version text NOT NULL,
    zero_digest text NOT NULL CHECK (zero_digest ~ '^[0-9a-f]{64}$'),
    blockers jsonb NOT NULL CHECK (jsonb_typeof(blockers) = 'array'),
    completed_at timestamptz NOT NULL,
    UNIQUE (operation_id, component_id, reference_epoch),
    FOREIGN KEY (component_id) REFERENCES iam_catalog_components(component_id) ON DELETE RESTRICT,
    CHECK (pg_catalog.uuid_extract_version(receipt_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(operation_id) = 7),
    CHECK (date_trunc('microseconds', completed_at) = completed_at)
);

GRANT SELECT ON iam_catalog_components TO forgecrud_runtime;

