-- +goose Up
CREATE TABLE analytics_saved_queries (
    id uuid NOT NULL,
    lifecycle_state text NOT NULL,
    row_revision bigint NOT NULL,
    reference_epoch bigint NOT NULL,
    current_draft_version_id uuid NOT NULL,
    created_by uuid NOT NULL,
    updated_by uuid NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT analytics_saved_queries_pkey PRIMARY KEY (id),
    CONSTRAINT analytics_saved_queries_draft_key UNIQUE (current_draft_version_id),
    CONSTRAINT analytics_saved_queries_identity_check CHECK (
        pg_catalog.uuid_extract_version(id) = 7
        AND pg_catalog.uuid_extract_version(current_draft_version_id) = 7
        AND created_by <> '00000000-0000-0000-0000-000000000000'::uuid
        AND updated_by <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT analytics_saved_queries_state_check CHECK (
        lifecycle_state IN ('active', 'retired')
        AND row_revision > 0 AND reference_epoch > 0
    ),
    CONSTRAINT analytics_saved_queries_time_check CHECK (
        pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)
        AND created_at <= updated_at
    )
);

CREATE TABLE analytics_saved_query_versions (
    id uuid NOT NULL,
    query_id uuid NOT NULL,
    version_no bigint NOT NULL,
    version_state text NOT NULL,
    row_revision bigint NOT NULL,
    name text NOT NULL,
    query_document bytea NOT NULL,
    query_sha256 bytea NOT NULL,
    parameter_document bytea NOT NULL,
    parameter_sha256 bytea NOT NULL,
    output_document bytea NOT NULL,
    output_sha256 bytea NOT NULL,
    manifest_sha256 bytea NOT NULL,
    cache_policy text NOT NULL,
    created_by uuid NOT NULL,
    updated_by uuid NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    published_by uuid,
    published_at timestamptz,
    CONSTRAINT analytics_saved_query_versions_pkey PRIMARY KEY (id),
    CONSTRAINT analytics_saved_query_versions_query_fk FOREIGN KEY (query_id)
        REFERENCES analytics_saved_queries (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_saved_query_versions_number_key UNIQUE (query_id, version_no),
    CONSTRAINT analytics_saved_query_versions_identity_key UNIQUE (id, query_id),
    CONSTRAINT analytics_saved_query_versions_identity_check CHECK (
        pg_catalog.uuid_extract_version(id) = 7 AND version_no > 0 AND row_revision > 0
        AND created_by <> '00000000-0000-0000-0000-000000000000'::uuid
        AND updated_by <> '00000000-0000-0000-0000-000000000000'::uuid
        AND pg_catalog.octet_length(name) BETWEEN 1 AND 100
    ),
    CONSTRAINT analytics_saved_query_versions_state_check CHECK (
        version_state IN ('draft', 'published', 'needs_revalidation', 'retired')
        AND cache_policy IN ('none', 'revision_bound')
        AND ((version_state = 'draft' AND published_by IS NULL AND published_at IS NULL)
          OR (version_state <> 'draft' AND published_by IS NOT NULL
              AND published_at IS NOT NULL))
    ),
    CONSTRAINT analytics_saved_query_versions_document_check CHECK (
        pg_catalog.octet_length(query_document) BETWEEN 2 AND 65536
        AND pg_catalog.sha256(query_document) = query_sha256
        AND pg_catalog.octet_length(parameter_document) BETWEEN 2 AND 65536
        AND pg_catalog.sha256(parameter_document) = parameter_sha256
        AND pg_catalog.octet_length(output_document) BETWEEN 2 AND 65536
        AND pg_catalog.sha256(output_document) = output_sha256
        AND pg_catalog.octet_length(manifest_sha256) = 32
        AND pg_catalog.pg_input_is_valid(pg_catalog.convert_from(query_document, 'UTF8'), 'jsonb')
        AND pg_catalog.pg_input_is_valid(pg_catalog.convert_from(parameter_document, 'UTF8'), 'jsonb')
        AND pg_catalog.pg_input_is_valid(pg_catalog.convert_from(output_document, 'UTF8'), 'jsonb')
    ),
    CONSTRAINT analytics_saved_query_versions_time_check CHECK (
        pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)
        AND created_at <= updated_at
        AND (published_at IS NULL OR (pg_catalog.isfinite(published_at)
             AND published_at >= created_at))
    )
);

ALTER TABLE analytics_saved_queries
    ADD CONSTRAINT analytics_saved_queries_current_draft_fk
    FOREIGN KEY (current_draft_version_id, id)
    REFERENCES analytics_saved_query_versions (id, query_id)
    ON UPDATE NO ACTION ON DELETE RESTRICT
    DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE analytics_query_dependencies (
    query_version_id uuid NOT NULL,
    ordinal smallint NOT NULL,
    owner_domain text NOT NULL,
    resource_kind text NOT NULL,
    resource_id uuid NOT NULL,
    resource_version_id uuid NOT NULL,
    owner_generation bigint NOT NULL,
    catalog_version bigint NOT NULL,
    query_contract_version bigint NOT NULL,
    policy_version bigint NOT NULL,
    data_revision_capability text NOT NULL,
    dependency_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT analytics_query_dependencies_pkey PRIMARY KEY (query_version_id, ordinal),
    CONSTRAINT analytics_query_dependencies_identity_key UNIQUE (
        query_version_id, owner_domain, resource_kind, resource_id, resource_version_id
    ),
    CONSTRAINT analytics_query_dependencies_version_fk FOREIGN KEY (query_version_id)
        REFERENCES analytics_saved_query_versions (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_query_dependencies_shape_check CHECK (
        ordinal BETWEEN 0 AND 31
        AND owner_domain IN ('iam')
        AND pg_catalog.octet_length(resource_kind) BETWEEN 1 AND 64
        AND resource_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND resource_version_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND owner_generation > 0 AND catalog_version > 0
        AND query_contract_version > 0 AND policy_version > 0
        AND pg_catalog.octet_length(data_revision_capability) BETWEEN 0 AND 128
        AND pg_catalog.octet_length(dependency_sha256) = 32
        AND pg_catalog.isfinite(created_at)
    )
);

CREATE INDEX analytics_saved_queries_list_idx
    ON analytics_saved_queries (id) WHERE lifecycle_state = 'active';

