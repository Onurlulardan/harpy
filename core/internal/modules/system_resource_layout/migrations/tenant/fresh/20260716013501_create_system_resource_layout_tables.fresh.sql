-- +goose Up
CREATE TABLE system_resource_layout_revisions (
    id uuid NOT NULL,
    resource_kind text NOT NULL,
    version bigint NOT NULL,
    state text NOT NULL,
    definition_json text NOT NULL,
    definition_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    published_at timestamptz,
    row_revision bigint NOT NULL,
    CONSTRAINT system_resource_layout_revisions_pkey PRIMARY KEY (id),
    CONSTRAINT system_resource_layout_revisions_resource_version_key
        UNIQUE (resource_kind, version),
    CONSTRAINT system_resource_layout_revisions_resource_id_key
        UNIQUE (resource_kind, id),
    CONSTRAINT system_resource_layout_revisions_id_check CHECK (
        id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT system_resource_layout_revisions_resource_check CHECK (
        resource_kind IN ('user', 'organization', 'role')
    ),
    CONSTRAINT system_resource_layout_revisions_version_check CHECK (version > 0),
    CONSTRAINT system_resource_layout_revisions_state_check CHECK (
        state IN ('draft', 'published')
    ),
    CONSTRAINT system_resource_layout_revisions_definition_check CHECK (
        system_resource_layout_definition_valid_v1(definition_json)
    ),
    CONSTRAINT system_resource_layout_revisions_digest_check CHECK (
        pg_catalog.octet_length(definition_sha256) = 32
        AND definition_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(definition_json, 'UTF8')
        )
    ),
    CONSTRAINT system_resource_layout_revisions_row_revision_check
        CHECK (row_revision > 0),
    CONSTRAINT system_resource_layout_revisions_timestamps_check CHECK (
        updated_at >= created_at
        AND (published_at IS NULL OR published_at >= created_at)
    ),
    CONSTRAINT system_resource_layout_revisions_state_timestamp_check CHECK (
        (state = 'draft' AND published_at IS NULL)
        OR (state = 'published' AND published_at IS NOT NULL)
    )
);

CREATE TABLE system_resource_layout_publications (
    resource_kind text NOT NULL,
    revision_id uuid NOT NULL,
    generation bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT system_resource_layout_publications_pkey
        PRIMARY KEY (resource_kind),
    CONSTRAINT system_resource_layout_publications_revision_key
        UNIQUE (revision_id),
    CONSTRAINT system_resource_layout_publications_resource_check CHECK (
        resource_kind IN ('user', 'organization', 'role')
    ),
    CONSTRAINT system_resource_layout_publications_generation_check
        CHECK (generation > 0),
    CONSTRAINT system_resource_layout_publications_timestamps_check
        CHECK (updated_at >= created_at),
    CONSTRAINT system_resource_layout_publications_revision_fk
        FOREIGN KEY (resource_kind, revision_id)
        REFERENCES system_resource_layout_revisions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

