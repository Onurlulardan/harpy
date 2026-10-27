-- +goose Up
CREATE TABLE catalog_menu_definitions (
    definition_id uuid PRIMARY KEY,
    revision bigint NOT NULL CHECK (revision > 0),
    active_version_id uuid,
    updated_at timestamptz NOT NULL,
    CHECK (date_trunc('microseconds', updated_at) = updated_at)
);

CREATE TABLE catalog_menu_versions (
    version_id uuid PRIMARY KEY,
    definition_id uuid NOT NULL,
    version_number bigint NOT NULL CHECK (version_number > 0),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL,
    UNIQUE (definition_id, version_number),
    CONSTRAINT catalog_menu_versions_definition_fk
        FOREIGN KEY (definition_id) REFERENCES catalog_menu_definitions(definition_id)
        ON DELETE RESTRICT,
    CHECK (date_trunc('microseconds', created_at) = created_at)
);

ALTER TABLE catalog_menu_definitions
ADD CONSTRAINT catalog_menu_definitions_active_version_fk
FOREIGN KEY (active_version_id) REFERENCES catalog_menu_versions(version_id)
ON DELETE RESTRICT;

CREATE TABLE catalog_menu_items (
    version_id uuid NOT NULL,
    item_id uuid NOT NULL,
    parent_item_id uuid,
    package_id uuid NOT NULL,
    item_kind text NOT NULL CHECK (item_kind IN (
        'GROUP','FORM','WORKFLOW','DASHBOARD','ENDPOINT','EXTERNAL_LINK'
    )),
    target_id uuid,
    external_url text,
    label text NOT NULL CHECK (length(label) BETWEEN 1 AND 128),
    sort_order integer NOT NULL CHECK (sort_order BETWEEN 0 AND 100000),
    PRIMARY KEY (version_id, item_id),
    CONSTRAINT catalog_menu_items_version_fk
        FOREIGN KEY (version_id) REFERENCES catalog_menu_versions(version_id)
        ON DELETE RESTRICT,
    CHECK ((item_kind = 'GROUP' AND target_id IS NULL AND external_url IS NULL)
        OR (item_kind = 'EXTERNAL_LINK' AND target_id IS NULL AND external_url LIKE 'https://%')
        OR (item_kind NOT IN ('GROUP','EXTERNAL_LINK') AND target_id IS NOT NULL AND external_url IS NULL))
);

CREATE TABLE catalog_menu_overrides (
    item_id uuid PRIMARY KEY,
    label text,
    sort_order integer,
    visible boolean,
    revision bigint NOT NULL CHECK (revision > 0),
    CHECK (label IS NULL OR length(label) BETWEEN 1 AND 128),
    CHECK (sort_order IS NULL OR sort_order BETWEEN 0 AND 100000)
);

CREATE TABLE catalog_menu_override_receipts (
    receipt_id uuid PRIMARY KEY,
    actor_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    request_sha256 text NOT NULL CHECK (request_sha256 ~ '^[0-9a-f]{64}$'),
    item_id uuid NOT NULL,
    result_revision bigint NOT NULL CHECK (result_revision > 0),
    result_label text,
    result_sort_order integer,
    result_visible boolean,
    completed_at timestamptz NOT NULL,
    UNIQUE (actor_id, idempotency_key),
    CHECK (result_label IS NULL OR length(result_label) BETWEEN 1 AND 128),
    CHECK (result_sort_order IS NULL OR result_sort_order BETWEEN 0 AND 100000),
    CHECK (date_trunc('microseconds', completed_at) = completed_at)
);
