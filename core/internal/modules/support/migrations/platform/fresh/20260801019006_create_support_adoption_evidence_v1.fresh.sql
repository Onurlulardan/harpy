-- +goose Up
-- +goose StatementBegin
ALTER TABLE platform.support_adoption_runs
    ADD COLUMN operation_id uuid NOT NULL,
    ADD COLUMN source_relation text NOT NULL,
    ADD COLUMN source_relation_oid oid NOT NULL,
    ADD COLUMN source_locator_sha256 bytea NOT NULL,
    ADD COLUMN source_manifest_sha256 bytea NOT NULL,
    ADD COLUMN transform_sha256 bytea NOT NULL,
    ADD COLUMN snapshot_id text NOT NULL,
    ADD COLUMN tenant_slug text NOT NULL DEFAULT '',
    ADD COLUMN tenant_name text NOT NULL DEFAULT '',
    ADD COLUMN expected_count bigint NOT NULL,
    ADD COLUMN expected_source_sha256 bytea NOT NULL,
    ADD COLUMN next_page bigint NOT NULL DEFAULT 1,
    ADD COLUMN lease_owner uuid,
    ADD COLUMN lease_claim_revision bigint,
    ADD COLUMN lease_generation bigint NOT NULL DEFAULT 0,
    ADD COLUMN lease_expires_at timestamp(6) with time zone,
    ADD COLUMN reconciled_at timestamp(6) with time zone,
    ADD COLUMN completed_at timestamp(6) with time zone,
    ADD CONSTRAINT support_adoption_runs_operation_key UNIQUE (operation_id),
    ADD CONSTRAINT support_adoption_runs_relation_check CHECK (
        pg_catalog.octet_length(source_relation) BETWEEN 1 AND 63
        AND source_relation_oid::bigint > 0
        AND source_schema ~ '^[a-z][a-z0-9_]{0,62}$'
    ),
    ADD CONSTRAINT support_adoption_runs_hash_check CHECK (
        pg_catalog.octet_length(source_locator_sha256) = 32
        AND pg_catalog.octet_length(source_manifest_sha256) = 32
        AND pg_catalog.octet_length(transform_sha256) = 32
        AND pg_catalog.octet_length(expected_source_sha256) = 32
    ),
    ADD CONSTRAINT support_adoption_runs_progress_check CHECK (
        expected_count >= 0 AND copied_count <= expected_count
        AND next_page > 0 AND lease_generation >= 0
        AND pg_catalog.octet_length(snapshot_id) BETWEEN 16 AND 200
    ),
    ADD CONSTRAINT support_adoption_runs_scope_check CHECK (
        (source_kind = 'contact_global' AND source_tenant_id IS NULL
            AND source_schema = 'public' AND source_relation = 'contact_messages'
            AND tenant_slug = '' AND tenant_name = '')
        OR
        (source_kind = 'bug_tenant' AND source_tenant_id IS NOT NULL
            AND source_schema NOT IN ('public','platform','pg_catalog','information_schema')
            AND source_schema !~ '^pg_' AND source_relation = 'bug_reports'
            AND pg_catalog.octet_length(tenant_slug) BETWEEN 1 AND 100
            AND pg_catalog.octet_length(tenant_name) BETWEEN 1 AND 200)
    ),
    ADD CONSTRAINT support_adoption_runs_lease_check CHECK (
        (lease_owner IS NULL) = (lease_expires_at IS NULL)
        AND (lease_owner IS NULL) = (lease_claim_revision IS NULL)
        AND (lease_claim_revision IS NULL OR lease_claim_revision > 0)
    );

ALTER TABLE platform.support_bug_reports
    ADD COLUMN legacy_screenshot_refs jsonb NOT NULL DEFAULT '[]'::jsonb
    CHECK (pg_catalog.jsonb_typeof(legacy_screenshot_refs) = 'array'
        AND pg_catalog.jsonb_array_length(legacy_screenshot_refs) <= 50
        AND pg_catalog.octet_length(legacy_screenshot_refs::text) <= 65536);

CREATE TABLE platform.support_adoption_mappings (
    run_id uuid NOT NULL REFERENCES platform.support_adoption_runs(run_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    source_kind text NOT NULL CHECK (source_kind IN ('contact_global','bug_tenant')),
    source_tenant_id uuid,
    source_id uuid NOT NULL,
    target_id uuid NOT NULL UNIQUE CHECK (pg_catalog.uuid_extract_version(target_id) = 7),
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256) = 32),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256) = 32),
    page_number bigint NOT NULL CHECK (page_number > 0),
    applied_at timestamp(6) with time zone NOT NULL,
    PRIMARY KEY (run_id, source_id),
    UNIQUE NULLS NOT DISTINCT (source_kind, source_tenant_id, source_id),
    CHECK ((source_kind = 'contact_global' AND source_tenant_id IS NULL)
        OR (source_kind = 'bug_tenant' AND source_tenant_id IS NOT NULL))
);

CREATE TABLE platform.support_adoption_page_receipts (
    run_id uuid NOT NULL REFERENCES platform.support_adoption_runs(run_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    page_number bigint NOT NULL CHECK (page_number > 0),
    cursor_before uuid,
    cursor_after uuid NOT NULL,
    row_count integer NOT NULL CHECK (row_count BETWEEN 1 AND 100),
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    source_page_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_page_sha256) = 32),
    target_page_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_page_sha256) = 32),
    state text NOT NULL CHECK (state IN ('applied','advanced')),
    applied_at timestamp(6) with time zone NOT NULL,
    advanced_at timestamp(6) with time zone,
    PRIMARY KEY (run_id, page_number),
    CHECK ((state = 'applied' AND advanced_at IS NULL)
        OR (state = 'advanced' AND advanced_at IS NOT NULL))
);

REVOKE ALL ON TABLE platform.support_adoption_runs FROM PUBLIC, harpy_runtime, harpy_adopter;
REVOKE ALL ON TABLE platform.support_adoption_mappings FROM PUBLIC, harpy_runtime, harpy_adopter;
REVOKE ALL ON TABLE platform.support_adoption_page_receipts FROM PUBLIC, harpy_runtime, harpy_adopter;
REVOKE ALL ON TABLE platform.support_bug_reports FROM harpy_adopter;
-- +goose StatementEnd
