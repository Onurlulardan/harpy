-- +goose Up
CREATE TABLE iam_profile_backfill_worker_capabilities (
    backend_pid integer NOT NULL,
    transaction_id xid8 NOT NULL,
    plan_id uuid NOT NULL,
    page_number integer NOT NULL,
    resource_kind text NOT NULL,
    resource_id uuid NOT NULL,
    actual_source_schema_id uuid NOT NULL,
    source_revision bigint NOT NULL,
    source_sha256 bytea NOT NULL,
    target_schema_id uuid NOT NULL,
    target_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_profile_backfill_worker_capabilities_pkey PRIMARY KEY (
        backend_pid, transaction_id, plan_id, resource_id
    ),
    CONSTRAINT iam_profile_backfill_worker_capabilities_page_fk
        FOREIGN KEY (plan_id, page_number)
        REFERENCES iam_profile_backfill_pages (plan_id, page_number)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_worker_capabilities_source_fk
        FOREIGN KEY (resource_kind, actual_source_schema_id)
        REFERENCES iam_profile_schema_versions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_worker_capabilities_target_fk
        FOREIGN KEY (resource_kind, target_schema_id)
        REFERENCES iam_profile_schema_versions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_backfill_worker_capabilities_kind_check
        CHECK (resource_kind IN ('user', 'organization', 'role')),
    CONSTRAINT iam_profile_backfill_worker_capabilities_value_check CHECK (
        backend_pid > 0 AND page_number > 0 AND source_revision > 0 AND
        actual_source_schema_id <> target_schema_id AND
        pg_catalog.octet_length(source_sha256) = 32 AND
        pg_catalog.octet_length(target_sha256) = 32
    )
);
