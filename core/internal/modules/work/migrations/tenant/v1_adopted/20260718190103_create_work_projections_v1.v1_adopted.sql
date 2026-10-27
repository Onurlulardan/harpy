-- +goose Up
CREATE TABLE work_resource_links (
    id uuid PRIMARY KEY,
    work_resource_kind text NOT NULL,
    work_resource_id uuid NOT NULL,
    target_owner text NOT NULL CHECK (target_owner IN ('FORMS','WORKFLOW')),
    target_kind text NOT NULL,
    target_id uuid NOT NULL,
    target_version_id uuid NOT NULL,
    usage text NOT NULL CHECK (usage IN ('READ','WRITE','SCHEMA','LIFECYCLE')),
    target_reference_epoch bigint NOT NULL CHECK (target_reference_epoch > 0),
    permission_revision bigint NOT NULL CHECK (permission_revision > 0),
    state text NOT NULL CHECK (state IN ('ACTIVE','RELEASED')),
    row_version bigint NOT NULL CHECK (row_version > 0),
    created_at timestamptz NOT NULL,
    released_at timestamptz,
    CHECK ((state = 'ACTIVE' AND released_at IS NULL) OR
           (state = 'RELEASED' AND released_at IS NOT NULL)),
    UNIQUE (work_resource_kind, work_resource_id, target_owner, target_kind,
            target_id, target_version_id, usage, id)
);

CREATE TABLE work_inbound_references (
    reference_id uuid PRIMARY KEY,
    target_kind text NOT NULL,
    target_id uuid NOT NULL,
    target_version_id uuid NOT NULL,
    source_module text NOT NULL,
    source_kind text NOT NULL,
    source_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    usage text NOT NULL CHECK (usage IN ('READ','WRITE','SCHEMA','LIFECYCLE')),
    state text NOT NULL CHECK (state IN ('ACTIVE','RELEASED')),
    created_at timestamptz NOT NULL,
    released_at timestamptz,
    UNIQUE (target_kind, target_id, target_version_id, source_module,
            source_kind, source_id, source_version_id, usage, reference_id),
    CHECK ((state = 'ACTIVE' AND released_at IS NULL) OR
           (state = 'RELEASED' AND released_at IS NOT NULL))
);

CREATE TABLE work_projection_streams (
    source_owner text NOT NULL CHECK (source_owner IN ('WORKFLOW','FORMS')),
    aggregate_kind text NOT NULL,
    aggregate_id uuid NOT NULL,
    through_sequence bigint NOT NULL CHECK (through_sequence >= 0),
    bootstrap_sequence bigint NOT NULL CHECK (bootstrap_sequence >= 0),
    state text NOT NULL CHECK (state IN ('CURRENT','GAP','REBUILDING','FAILED')),
    state_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(state_sha256) = 32),
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (source_owner, aggregate_kind, aggregate_id)
);

CREATE TABLE work_projection_receipts (
    consumer_name text NOT NULL,
    event_id uuid NOT NULL,
    source_owner text NOT NULL CHECK (source_owner IN ('WORKFLOW','FORMS')),
    aggregate_kind text NOT NULL,
    aggregate_id uuid NOT NULL,
    event_sequence bigint NOT NULL CHECK (event_sequence > 0),
    event_version bigint NOT NULL CHECK (event_version > 0),
    payload_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(payload_sha256) = 32),
    applied_at timestamptz NOT NULL,
    PRIMARY KEY (consumer_name, event_id)
);

CREATE TABLE work_inbox_items (
    id uuid PRIMARY KEY,
    source_owner text NOT NULL CHECK (source_owner IN ('WORK','WORKFLOW','FORMS')),
    source_kind text NOT NULL,
    source_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    source_revision bigint NOT NULL CHECK (source_revision > 0),
    permission_revision bigint NOT NULL CHECK (permission_revision > 0),
    event_sequence bigint NOT NULL CHECK (event_sequence > 0),
    summary text NOT NULL CHECK (pg_catalog.octet_length(summary) <= 2000),
    due_at timestamptz,
    state text NOT NULL CHECK (state IN ('VISIBLE','STALE','REVOKED')),
    updated_at timestamptz NOT NULL,
    UNIQUE (source_owner, source_kind, source_id, source_version_id)
);

CREATE TABLE work_calendar_items (
    id uuid PRIMARY KEY,
    source_owner text NOT NULL CHECK (source_owner IN ('WORK','WORKFLOW','FORMS')),
    source_kind text NOT NULL,
    source_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    source_revision bigint NOT NULL CHECK (source_revision > 0),
    permission_revision bigint NOT NULL CHECK (permission_revision > 0),
    event_sequence bigint NOT NULL CHECK (event_sequence > 0),
    summary text NOT NULL CHECK (pg_catalog.octet_length(summary) <= 2000),
    starts_at timestamptz,
    ends_at timestamptz,
    start_date date,
    end_date date,
    state text NOT NULL CHECK (state IN ('VISIBLE','STALE','REVOKED')),
    updated_at timestamptz NOT NULL,
    UNIQUE (source_owner, source_kind, source_id, source_version_id),
    CHECK ((starts_at IS NOT NULL AND ends_at > starts_at AND start_date IS NULL AND end_date IS NULL) OR
           (starts_at IS NULL AND ends_at IS NULL AND start_date IS NOT NULL AND end_date > start_date))
);

CREATE INDEX work_inbox_items_due_idx ON work_inbox_items(due_at, id) WHERE state = 'VISIBLE';
CREATE INDEX work_calendar_items_time_idx ON work_calendar_items(starts_at, id) WHERE state = 'VISIBLE';
