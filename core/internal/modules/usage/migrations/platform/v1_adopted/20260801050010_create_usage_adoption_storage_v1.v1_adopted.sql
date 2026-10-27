-- +goose Up
-- +goose StatementBegin
CREATE TABLE platform.usage_adoption_runs (
    run_id uuid PRIMARY KEY CHECK (pg_catalog.uuid_extract_version(run_id) = 7),
    operation_id uuid NOT NULL UNIQUE,
    source_kind text NOT NULL CHECK (source_kind IN ('zero','legacy_trace_logs')),
    source_schema text NOT NULL CHECK (source_schema ~ '^[a-z][a-z0-9_]{0,62}$'),
    source_relation text NOT NULL CHECK (source_relation ~ '^[a-z][a-z0-9_]{0,62}$'),
    source_relation_oid oid NOT NULL,
    source_count bigint NOT NULL CHECK (source_count >= 0),
    source_root_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_root_sha256) = 32),
    snapshot_id text NOT NULL CHECK (pg_catalog.octet_length(snapshot_id) BETWEEN 16 AND 200),
    transform_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(transform_sha256) = 32),
    next_page bigint NOT NULL DEFAULT 1 CHECK (next_page > 0),
    cursor_id uuid,
    copied_count bigint NOT NULL DEFAULT 0 CHECK (copied_count >= 0),
    status text NOT NULL CHECK (status IN ('planned','applying','reconciling','completed','fenced')),
    lease_owner uuid,
    lease_generation bigint NOT NULL DEFAULT 0 CHECK (lease_generation >= 0),
    lease_expires_at timestamp(6) with time zone,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamp(6) with time zone NOT NULL,
    updated_at timestamp(6) with time zone NOT NULL,
    completed_at timestamp(6) with time zone,
    CHECK ((source_kind = 'zero' AND source_relation_oid = 0 AND source_count = 0)
        OR (source_kind = 'legacy_trace_logs' AND source_relation_oid::bigint > 0)),
    CHECK ((lease_owner IS NULL) = (lease_expires_at IS NULL)),
    CHECK (copied_count <= source_count),
    CHECK ((status = 'completed') = (completed_at IS NOT NULL))
);

CREATE TABLE platform.usage_adoption_mappings (
    run_id uuid NOT NULL REFERENCES platform.usage_adoption_runs(run_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    source_id uuid NOT NULL,
    target_id uuid NOT NULL UNIQUE CHECK (pg_catalog.uuid_extract_version(target_id) = 7),
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256) = 32),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256) = 32),
    page_number bigint NOT NULL CHECK (page_number > 0),
    applied_at timestamp(6) with time zone NOT NULL,
    PRIMARY KEY (run_id, source_id)
);

CREATE TABLE platform.usage_adoption_page_receipts (
    run_id uuid NOT NULL REFERENCES platform.usage_adoption_runs(run_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    page_number bigint NOT NULL CHECK (page_number > 0),
    cursor_before uuid,
    cursor_after uuid NOT NULL,
    row_count integer NOT NULL CHECK (row_count BETWEEN 1 AND 100),
    source_page_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_page_sha256) = 32),
    target_page_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_page_sha256) = 32),
    state text NOT NULL CHECK (state IN ('applied','advanced')),
    applied_at timestamp(6) with time zone NOT NULL,
    advanced_at timestamp(6) with time zone,
    PRIMARY KEY (run_id, page_number),
    CHECK ((state = 'applied' AND advanced_at IS NULL)
        OR (state = 'advanced' AND advanced_at IS NOT NULL))
);

-- Adoption evidence remains owner-only; the adopter is granted only the
-- SECURITY DEFINER routines declared by the final grant migration.
-- +goose StatementEnd
