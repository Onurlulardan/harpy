-- +goose Up
CREATE TABLE documents_folders (
    id uuid PRIMARY KEY,
    parent_id uuid,
    owner_id uuid NOT NULL,
    display_name text NOT NULL
        CHECK (pg_catalog.octet_length(display_name) BETWEEN 1 AND 1024),
    lifecycle_state text NOT NULL
        CHECK (lifecycle_state IN ('ACTIVE','TRASHED')),
    row_version bigint NOT NULL CHECK (row_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    trashed_at timestamptz,
    CONSTRAINT documents_folders_parent_fk FOREIGN KEY (parent_id)
        REFERENCES documents_folders(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (parent_id IS NULL OR parent_id <> id),
    CHECK (pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)),
    CHECK ((lifecycle_state = 'ACTIVE' AND trashed_at IS NULL) OR
           (lifecycle_state = 'TRASHED' AND trashed_at IS NOT NULL))
);

CREATE TABLE documents_items (
    id uuid PRIMARY KEY,
    folder_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    current_version_id uuid,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'ACTIVE','TRASHED','PURGE_COOLING','PURGING','PURGED'
    )),
    resume_state text CHECK (resume_state IN ('ACTIVE','TRASHED')),
    row_version bigint NOT NULL CHECK (row_version > 0),
    trashed_at timestamptz,
    retention_ends_at timestamptz,
    purge_requested_at timestamptz,
    purge_starts_at timestamptz,
    purge_reason text CHECK (
        purge_reason IS NULL OR pg_catalog.octet_length(purge_reason) BETWEEN 3 AND 2048
    ),
    purged_at timestamptz,
    tombstone_until timestamptz,
    purge_evidence_sha256 bytea CHECK (
        purge_evidence_sha256 IS NULL OR pg_catalog.octet_length(purge_evidence_sha256) = 32
    ),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT documents_items_folder_fk FOREIGN KEY (folder_id)
        REFERENCES documents_folders(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)),
    CHECK ((lifecycle_state = 'ACTIVE' AND trashed_at IS NULL AND
            retention_ends_at IS NULL) OR lifecycle_state <> 'ACTIVE'),
    CHECK ((lifecycle_state = 'TRASHED' AND trashed_at IS NOT NULL AND
            retention_ends_at > trashed_at) OR lifecycle_state <> 'TRASHED'),
    CHECK ((lifecycle_state = 'PURGE_COOLING' AND resume_state IS NOT NULL AND
            purge_requested_at IS NOT NULL AND purge_starts_at > purge_requested_at) OR
           lifecycle_state <> 'PURGE_COOLING'),
    CHECK ((lifecycle_state = 'PURGED' AND current_version_id IS NULL AND
            purged_at IS NOT NULL AND tombstone_until > purged_at AND
            purge_evidence_sha256 IS NOT NULL) OR lifecycle_state <> 'PURGED')
);

CREATE INDEX documents_folders_parent_list_idx
    ON documents_folders(parent_id, display_name, id) WHERE lifecycle_state = 'ACTIVE';
CREATE INDEX documents_items_folder_list_idx
    ON documents_items(folder_id, updated_at DESC, id) WHERE lifecycle_state <> 'PURGED';
CREATE INDEX documents_items_retention_idx
    ON documents_items(retention_ends_at, id) WHERE lifecycle_state = 'TRASHED';
CREATE INDEX documents_items_purge_cooling_idx
    ON documents_items(purge_starts_at, id) WHERE lifecycle_state = 'PURGE_COOLING';

