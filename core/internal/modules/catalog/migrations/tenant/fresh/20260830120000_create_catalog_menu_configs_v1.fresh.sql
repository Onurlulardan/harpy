-- +goose Up
CREATE TABLE catalog_menu_configs (
    config_id uuid PRIMARY KEY,
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 128),
    config jsonb NOT NULL CHECK (jsonb_typeof(config) = 'object'),
    is_active boolean NOT NULL,
    created_by uuid NOT NULL,
    updated_by uuid NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CHECK (config_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (created_by <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (updated_by <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (date_trunc('microseconds', created_at) = created_at),
    CHECK (date_trunc('microseconds', updated_at) = updated_at),
    CHECK (updated_at >= created_at)
);

CREATE UNIQUE INDEX catalog_menu_configs_one_active
    ON catalog_menu_configs (is_active)
    WHERE is_active;

GRANT SELECT ON TABLE catalog_menu_configs TO harpy_runtime;
GRANT INSERT (
    config_id, name, config, is_active, created_by, updated_by, created_at, updated_at
) ON catalog_menu_configs TO harpy_runtime;
GRANT UPDATE (name, config, is_active, updated_by, updated_at)
    ON catalog_menu_configs TO harpy_runtime;
