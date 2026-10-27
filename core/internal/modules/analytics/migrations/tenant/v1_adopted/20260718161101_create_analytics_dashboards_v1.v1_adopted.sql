-- +goose Up
CREATE TABLE analytics_dashboards (
    id uuid NOT NULL,
    lifecycle_state text NOT NULL,
    row_revision bigint NOT NULL,
    reference_epoch bigint NOT NULL,
    current_draft_version_id uuid NOT NULL,
    created_by uuid NOT NULL,
    updated_by uuid NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT analytics_dashboards_pkey PRIMARY KEY (id),
    CONSTRAINT analytics_dashboards_draft_key UNIQUE (current_draft_version_id),
    CONSTRAINT analytics_dashboards_identity_check CHECK (
        pg_catalog.uuid_extract_version(id) = 7
        AND pg_catalog.uuid_extract_version(current_draft_version_id) = 7
        AND created_by <> '00000000-0000-0000-0000-000000000000'::uuid
        AND updated_by <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT analytics_dashboards_state_check CHECK (
        lifecycle_state IN ('active', 'retired')
        AND row_revision > 0 AND reference_epoch > 0
    ),
    CONSTRAINT analytics_dashboards_time_check CHECK (
        pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)
        AND created_at <= updated_at
    )
);

CREATE TABLE analytics_dashboard_versions (
    id uuid NOT NULL,
    dashboard_id uuid NOT NULL,
    version_no bigint NOT NULL,
    version_state text NOT NULL,
    row_revision bigint NOT NULL,
    name text NOT NULL,
    visibility text NOT NULL,
    layout_document bytea NOT NULL,
    layout_sha256 bytea NOT NULL,
    manifest_sha256 bytea NOT NULL,
    created_by uuid NOT NULL,
    updated_by uuid NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    published_by uuid,
    published_at timestamptz,
    CONSTRAINT analytics_dashboard_versions_pkey PRIMARY KEY (id),
    CONSTRAINT analytics_dashboard_versions_dashboard_fk FOREIGN KEY (dashboard_id)
        REFERENCES analytics_dashboards (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_versions_number_key UNIQUE (dashboard_id, version_no),
    CONSTRAINT analytics_dashboard_versions_identity_key UNIQUE (id, dashboard_id),
    CONSTRAINT analytics_dashboard_versions_shape_check CHECK (
        pg_catalog.uuid_extract_version(id) = 7 AND version_no > 0 AND row_revision > 0
        AND version_state IN ('draft', 'published', 'retired')
        AND visibility IN ('private', 'tenant')
        AND pg_catalog.octet_length(name) BETWEEN 1 AND 100
        AND pg_catalog.octet_length(layout_document) BETWEEN 2 AND 65536
        AND pg_catalog.sha256(layout_document) = layout_sha256
        AND pg_catalog.octet_length(manifest_sha256) = 32
        AND pg_catalog.pg_input_is_valid(pg_catalog.convert_from(layout_document, 'UTF8'), 'jsonb')
        AND ((version_state = 'draft' AND published_by IS NULL AND published_at IS NULL)
          OR (version_state <> 'draft' AND published_by IS NOT NULL AND published_at IS NOT NULL))
    ),
    CONSTRAINT analytics_dashboard_versions_time_check CHECK (
        pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)
        AND created_at <= updated_at
        AND (published_at IS NULL OR (pg_catalog.isfinite(published_at)
             AND published_at >= created_at))
    )
);

ALTER TABLE analytics_dashboards
    ADD CONSTRAINT analytics_dashboards_current_draft_fk
    FOREIGN KEY (current_draft_version_id, id)
    REFERENCES analytics_dashboard_versions (id, dashboard_id)
    ON UPDATE NO ACTION ON DELETE RESTRICT
    DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE analytics_dashboard_widgets (
    dashboard_version_id uuid NOT NULL,
    widget_id uuid NOT NULL,
    position smallint NOT NULL,
    query_version_id uuid NOT NULL,
    parameter_document bytea NOT NULL,
    parameter_sha256 bytea NOT NULL,
    output_document bytea NOT NULL,
    output_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT analytics_dashboard_widgets_pkey PRIMARY KEY (
        dashboard_version_id, widget_id
    ),
    CONSTRAINT analytics_dashboard_widgets_position_key UNIQUE (
        dashboard_version_id, position
    ),
    CONSTRAINT analytics_dashboard_widgets_version_fk FOREIGN KEY (dashboard_version_id)
        REFERENCES analytics_dashboard_versions (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_widgets_shape_check CHECK (
        pg_catalog.uuid_extract_version(widget_id) = 7
        AND pg_catalog.uuid_extract_version(query_version_id) = 7
        AND position BETWEEN 0 AND 29
        AND pg_catalog.octet_length(parameter_document) BETWEEN 2 AND 65536
        AND pg_catalog.sha256(parameter_document) = parameter_sha256
        AND pg_catalog.octet_length(output_document) BETWEEN 2 AND 65536
        AND pg_catalog.sha256(output_document) = output_sha256
        AND pg_catalog.isfinite(created_at)
    )
);

CREATE TABLE analytics_dashboard_acls (
    dashboard_id uuid NOT NULL,
    principal_kind text NOT NULL,
    principal_id uuid NOT NULL,
    access_level text NOT NULL,
    row_revision bigint NOT NULL,
    granted_by uuid NOT NULL,
    granted_at timestamptz NOT NULL,
    CONSTRAINT analytics_dashboard_acls_pkey PRIMARY KEY (
        dashboard_id, principal_kind, principal_id
    ),
    CONSTRAINT analytics_dashboard_acls_dashboard_fk FOREIGN KEY (dashboard_id)
        REFERENCES analytics_dashboards (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_acls_shape_check CHECK (
        principal_kind IN ('user', 'role')
        AND principal_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND access_level IN ('read', 'edit', 'publish')
        AND row_revision > 0
        AND granted_by <> '00000000-0000-0000-0000-000000000000'::uuid
        AND pg_catalog.isfinite(granted_at)
    )
);

CREATE TABLE analytics_dashboard_preferences (
    principal_id uuid NOT NULL,
    surface text NOT NULL,
    dashboard_id uuid NOT NULL,
    row_revision bigint NOT NULL,
    updated_by uuid NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT analytics_dashboard_preferences_pkey PRIMARY KEY (principal_id, surface),
    CONSTRAINT analytics_dashboard_preferences_dashboard_fk FOREIGN KEY (dashboard_id)
        REFERENCES analytics_dashboards (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_preferences_shape_check CHECK (
        principal_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND surface ~ '^[a-z][a-z0-9_.-]{0,63}$'
        AND row_revision > 0
        AND updated_by <> '00000000-0000-0000-0000-000000000000'::uuid
        AND pg_catalog.isfinite(updated_at)
    )
);

CREATE INDEX analytics_dashboards_list_idx
    ON analytics_dashboards (id) WHERE lifecycle_state = 'active';

