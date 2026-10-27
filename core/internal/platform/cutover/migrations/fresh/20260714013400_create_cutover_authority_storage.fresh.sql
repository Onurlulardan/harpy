-- +goose Up
CREATE SCHEMA cutover_authority;
REVOKE ALL ON SCHEMA cutover_authority FROM PUBLIC;

CREATE TABLE cutover_authority.authority_adoption_inventory (
    catalog_generation integer PRIMARY KEY CHECK (catalog_generation > 0),
    operation_id uuid NOT NULL UNIQUE,
    receipt_id uuid NOT NULL UNIQUE,
    receipt_sha256 text NOT NULL UNIQUE CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    receipt_document bytea NOT NULL
        CHECK (pg_catalog.octet_length(receipt_document) BETWEEN 2 AND 16777216),
    operation_sha256 text NOT NULL CHECK (operation_sha256 ~ '^[0-9a-f]{64}$'),
    release_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    database_origin text NOT NULL CHECK (database_origin IN ('fresh', 'v1_adopted')),
    system_identifier numeric(20, 0) NOT NULL CHECK (system_identifier > 0),
    database_oid oid NOT NULL,
    catalog_contract_sha256 text NOT NULL
        CHECK (catalog_contract_sha256 ~ '^[0-9a-f]{64}$'),
    catalog_instance_sha256 text NOT NULL
        CHECK (catalog_instance_sha256 ~ '^[0-9a-f]{64}$'),
    dependency_contract_sha256 text NOT NULL
        CHECK (dependency_contract_sha256 ~ '^[0-9a-f]{64}$'),
    dependency_instance_sha256 text NOT NULL
        CHECK (dependency_instance_sha256 ~ '^[0-9a-f]{64}$'),
    object_count integer NOT NULL CHECK (object_count BETWEEN 18 AND 256),
    dependency_count integer NOT NULL CHECK (dependency_count = 19),
    adopted_at timestamptz NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    CHECK (receipt_sha256 = pg_catalog.encode(pg_catalog.sha256(receipt_document), 'hex'))
);

CREATE TABLE cutover_authority.finalization_authorizations (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    finalizer_lease_id uuid NOT NULL UNIQUE,
    finalizer_lease_sha256 text NOT NULL,
    finalizer_lease_document bytea NOT NULL,
    adoption_lease_sha256 text NOT NULL,
    adoption_lease_document bytea NOT NULL,
    seal_receipt_sha256 text NOT NULL,
    seal_receipt_document bytea NOT NULL,
    cleanup_receipt_sha256 text NOT NULL,
    cleanup_receipt_document bytea NOT NULL,
    authority_receipt_sha256 text NOT NULL,
    authority_receipt_document bytea NOT NULL,
    terminal_state text NOT NULL CHECK (terminal_state IN ('completed','failed','cancelled')),
    terminal_evidence_sha256 text NOT NULL CHECK (terminal_evidence_sha256 ~ '^[0-9a-f]{64}$'),
    terminal_evidence_document bytea NOT NULL,
    finalization_state_sha256 text NOT NULL CHECK (finalization_state_sha256 ~ '^[0-9a-f]{64}$'),
    finalizer_expires_at timestamptz NOT NULL,
    authorizer_valid_until timestamptz NOT NULL,
    authorized_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    authorized_by text NOT NULL DEFAULT SESSION_USER
        CHECK (authorized_by = 'forgecrud_cluster_admin'),
    PRIMARY KEY (operation_id, run_id, finalizer_lease_id),
    CHECK (finalizer_lease_sha256 =
        pg_catalog.encode(pg_catalog.sha256(finalizer_lease_document), 'hex')),
    CHECK (adoption_lease_sha256 =
        pg_catalog.encode(pg_catalog.sha256(adoption_lease_document), 'hex')),
    CHECK (seal_receipt_sha256 =
        pg_catalog.encode(pg_catalog.sha256(seal_receipt_document), 'hex')),
    CHECK (cleanup_receipt_sha256 =
        pg_catalog.encode(pg_catalog.sha256(cleanup_receipt_document), 'hex')),
    CHECK (authority_receipt_sha256 =
        pg_catalog.encode(pg_catalog.sha256(authority_receipt_document), 'hex')),
    CHECK (terminal_evidence_sha256 =
        pg_catalog.encode(pg_catalog.sha256(terminal_evidence_document), 'hex')),
    CHECK (pg_catalog.octet_length(finalizer_lease_document) BETWEEN 2 AND 65536),
    CHECK (pg_catalog.octet_length(adoption_lease_document) BETWEEN 2 AND 16777216),
    CHECK (pg_catalog.octet_length(seal_receipt_document) BETWEEN 2 AND 65536),
    CHECK (pg_catalog.octet_length(cleanup_receipt_document) BETWEEN 2 AND 65536),
    CHECK (pg_catalog.octet_length(authority_receipt_document) BETWEEN 2 AND 16777216),
    CHECK (pg_catalog.octet_length(terminal_evidence_document) BETWEEN 2 AND 16777216)
);

CREATE TABLE cutover_authority.finalization_receipts (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    finalizer_lease_id uuid NOT NULL UNIQUE,
    finalizer_lease_sha256 text NOT NULL,
    finalizer_lease_document bytea NOT NULL,
    adoption_lease_sha256 text NOT NULL,
    adoption_lease_document bytea NOT NULL,
    seal_receipt_sha256 text NOT NULL,
    seal_receipt_document bytea NOT NULL,
    cleanup_receipt_sha256 text NOT NULL,
    cleanup_receipt_document bytea NOT NULL,
    authority_receipt_sha256 text NOT NULL,
    authority_receipt_document bytea NOT NULL,
    terminal_state text NOT NULL CHECK (terminal_state IN ('completed','failed','cancelled')),
    terminal_evidence_sha256 text NOT NULL CHECK (terminal_evidence_sha256 ~ '^[0-9a-f]{64}$'),
    terminal_evidence_document bytea NOT NULL,
    finalization_state_sha256 text NOT NULL CHECK (finalization_state_sha256 ~ '^[0-9a-f]{64}$'),
    tenant_requirement_catalog_sha256 text,
    tenant_requirement_catalog_document bytea,
    tenant_requirement_count integer,
    tenant_target_count integer,
    finalized_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    finalized_by text NOT NULL DEFAULT SESSION_USER
        CHECK (finalized_by = 'forgecrud_cutover_finalizer'),
    PRIMARY KEY (operation_id, run_id),
    CHECK (finalizer_lease_sha256 =
        pg_catalog.encode(pg_catalog.sha256(finalizer_lease_document), 'hex')),
    CHECK (adoption_lease_sha256 =
        pg_catalog.encode(pg_catalog.sha256(adoption_lease_document), 'hex')),
    CHECK (seal_receipt_sha256 =
        pg_catalog.encode(pg_catalog.sha256(seal_receipt_document), 'hex')),
    CHECK (cleanup_receipt_sha256 =
        pg_catalog.encode(pg_catalog.sha256(cleanup_receipt_document), 'hex')),
    CHECK (authority_receipt_sha256 =
        pg_catalog.encode(pg_catalog.sha256(authority_receipt_document), 'hex')),
    CHECK (terminal_evidence_sha256 =
        pg_catalog.encode(pg_catalog.sha256(terminal_evidence_document), 'hex')),
    CHECK ((terminal_state = 'completed' AND
            tenant_requirement_catalog_sha256 IS NOT NULL AND
            tenant_requirement_catalog_sha256 ~ '^[0-9a-f]{64}$' AND
            tenant_requirement_catalog_document IS NOT NULL AND
            tenant_requirement_count IS NOT NULL AND tenant_requirement_count BETWEEN 1 AND 10000 AND
            tenant_target_count IS NOT NULL AND tenant_target_count BETWEEN 0 AND 10000) OR
           (terminal_state IN ('failed','cancelled') AND
            tenant_requirement_catalog_sha256 IS NULL AND
            tenant_requirement_catalog_document IS NULL AND
            tenant_requirement_count IS NULL AND tenant_target_count IS NULL)),
    CHECK (tenant_requirement_catalog_sha256 IS NULL OR
        tenant_requirement_catalog_sha256 = pg_catalog.encode(
            pg_catalog.sha256(tenant_requirement_catalog_document), 'hex')),
    CHECK (tenant_requirement_catalog_document IS NULL OR
        pg_catalog.octet_length(tenant_requirement_catalog_document) BETWEEN 2 AND 16777216),
    CHECK (pg_catalog.octet_length(finalizer_lease_document) BETWEEN 2 AND 65536),
    CHECK (pg_catalog.octet_length(adoption_lease_document) BETWEEN 2 AND 16777216),
    CHECK (pg_catalog.octet_length(seal_receipt_document) BETWEEN 2 AND 65536),
    CHECK (pg_catalog.octet_length(cleanup_receipt_document) BETWEEN 2 AND 65536),
    CHECK (pg_catalog.octet_length(authority_receipt_document) BETWEEN 2 AND 16777216),
    CHECK (pg_catalog.octet_length(terminal_evidence_document) BETWEEN 2 AND 16777216)
);

REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM forgecrud_migrator;
REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM forgecrud_cluster_admin;
REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM forgecrud_cutover_finalizer;
GRANT USAGE ON SCHEMA cutover_authority TO forgecrud_cluster_admin;
GRANT USAGE ON SCHEMA cutover_authority TO forgecrud_cutover_finalizer;
GRANT USAGE ON SCHEMA platform TO forgecrud_cutover_authority_owner;
GRANT SELECT ON platform.cutover_operations,
    platform.cutover_operation_relations, platform.cutover_operation_checkpoint_keys,
    platform.cutover_operation_columns,
    platform.cutover_operation_procedures, platform.cutover_initial_copy_receipts,
    platform.cutover_batch_receipts, platform.cutover_copy_checkpoints,
    platform.applied_change_receipts, platform.cutover_change_journal,
    platform.cutover_credential_receipts,
    platform.tenant_migration_requirement_catalogs,
    platform.tenant_migration_requirements,
    platform.tenant_migration_requirement_targets,
    platform.tenant_migration_checkpoints, platform.tenants,
    platform.release_generations, platform.schema_contracts
    TO forgecrud_cutover_authority_owner;
GRANT MAINTAIN ON platform.cutover_operations,
    platform.cutover_operation_relations, platform.cutover_operation_checkpoint_keys,
    platform.cutover_operation_columns,
    platform.cutover_operation_procedures, platform.cutover_initial_copy_receipts,
    platform.cutover_batch_receipts, platform.cutover_copy_checkpoints,
    platform.applied_change_receipts, platform.cutover_change_journal,
    platform.cutover_credential_receipts,
    platform.tenant_migration_requirement_catalogs,
    platform.tenant_migration_requirements,
    platform.tenant_migration_requirement_targets,
    platform.tenant_migration_checkpoints, platform.tenants,
    platform.release_generations, platform.schema_contracts
    TO forgecrud_cutover_authority_owner;
GRANT UPDATE (state, blocked_reason, updated_at) ON platform.cutover_operations
    TO forgecrud_cutover_authority_owner;
GRANT INSERT (receipt_id,operation_id,run_id,lease_id,receipt_kind,contract_version,
    lease_manifest_sha256,seal_receipt_sha256,receipt_sha256,canonical_document,
    issued_at,verified_at) ON platform.cutover_credential_receipts
    TO forgecrud_cutover_authority_owner;
