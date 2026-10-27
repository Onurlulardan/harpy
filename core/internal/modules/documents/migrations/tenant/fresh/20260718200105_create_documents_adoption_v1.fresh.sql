-- +goose Up
CREATE TABLE documents_adoption_runs (
    run_id uuid PRIMARY KEY,
    cutover_generation bigint NOT NULL CHECK (cutover_generation > 0),
    manifest_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(manifest_sha256) = 32),
    transform_version text NOT NULL CHECK (transform_version = 'documents_adoption_v1'),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'INVENTORY','COPYING','SHADOW','RECONCILING','QUIESCED',
        'ACCEPTED','BLOCKED','REVERSED'
    )),
    source_count bigint NOT NULL CHECK (source_count >= 0),
    target_count bigint NOT NULL CHECK (target_count >= 0),
    source_plaintext_bytes bigint NOT NULL CHECK (source_plaintext_bytes >= 0),
    target_plaintext_bytes bigint NOT NULL CHECK (target_plaintext_bytes >= 0),
    source_merkle_sha256 bytea CHECK (
        source_merkle_sha256 IS NULL OR pg_catalog.octet_length(source_merkle_sha256) = 32
    ),
    target_merkle_sha256 bytea CHECK (
        target_merkle_sha256 IS NULL OR pg_catalog.octet_length(target_merkle_sha256) = 32
    ),
    pending_object_count bigint NOT NULL CHECK (pending_object_count >= 0),
    pending_job_count bigint NOT NULL CHECK (pending_job_count >= 0),
    final_lsn pg_lsn,
    applied_lsn pg_lsn,
    acceptance_sha256 bytea CHECK (
        acceptance_sha256 IS NULL OR pg_catalog.octet_length(acceptance_sha256) = 32
    ),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CHECK (applied_lsn IS NULL OR final_lsn IS NULL OR applied_lsn <= final_lsn),
    CHECK (lifecycle_state <> 'ACCEPTED' OR
           (source_count = target_count AND source_plaintext_bytes = target_plaintext_bytes AND
            source_merkle_sha256 = target_merkle_sha256 AND pending_object_count = 0 AND
            pending_job_count = 0 AND final_lsn = applied_lsn AND acceptance_sha256 IS NOT NULL))
);

CREATE TABLE documents_blob_copy_checkpoints (
    run_id uuid NOT NULL,
    source_document_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    source_object_key text NOT NULL,
    source_object_revision text NOT NULL,
    source_bytes bigint NOT NULL CHECK (source_bytes >= 0),
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256) = 32),
    target_document_id uuid NOT NULL,
    target_version_id uuid NOT NULL,
    target_object_key text,
    target_object_version text,
    target_ciphertext_bytes bigint,
    target_plaintext_sha256 bytea CHECK (
        target_plaintext_sha256 IS NULL OR pg_catalog.octet_length(target_plaintext_sha256) = 32
    ),
    metadata_sha256 bytea CHECK (
        metadata_sha256 IS NULL OR pg_catalog.octet_length(metadata_sha256) = 32
    ),
    applied_lsn pg_lsn,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'INVENTORIED','COPYING','VERIFIED','BLOCKED','SOURCE_HOLD_RELEASED'
    )),
    issue_code text,
    verified_at timestamptz,
    source_hold_released_at timestamptz,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (run_id, source_version_id, source_object_revision),
    CONSTRAINT documents_blob_copy_checkpoints_run_fk FOREIGN KEY (run_id)
        REFERENCES documents_adoption_runs(run_id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    UNIQUE (run_id, target_version_id),
    CHECK ((lifecycle_state IN ('VERIFIED','SOURCE_HOLD_RELEASED')) =
           (target_object_key IS NOT NULL AND target_object_version IS NOT NULL AND
            target_ciphertext_bytes IS NOT NULL AND target_plaintext_sha256 = source_sha256 AND
            metadata_sha256 IS NOT NULL AND applied_lsn IS NOT NULL AND verified_at IS NOT NULL)),
    CHECK ((lifecycle_state = 'BLOCKED') = (issue_code IS NOT NULL)),
    CHECK ((lifecycle_state = 'SOURCE_HOLD_RELEASED') =
           (source_hold_released_at IS NOT NULL))
);

CREATE TABLE documents_adoption_issues (
    run_id uuid NOT NULL,
    issue_id uuid NOT NULL,
    source_version_id uuid,
    issue_code text NOT NULL CHECK (issue_code IN (
        'MISSING_OBJECT','COLLISION','SIZE_MISMATCH','CHECKSUM_MISMATCH',
        'VERSION_CHAIN_MISMATCH','OBJECT_OUTCOME_UNKNOWN','PENDING_JOB'
    )),
    evidence_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(evidence_sha256) = 32),
    created_at timestamptz NOT NULL,
    PRIMARY KEY (run_id, issue_id),
    CONSTRAINT documents_adoption_issues_run_fk FOREIGN KEY (run_id)
        REFERENCES documents_adoption_runs(run_id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE documents_compatibility_hits (
    route_contract text NOT NULL,
    consumer_id text NOT NULL,
    build_id text NOT NULL,
    hit_count bigint NOT NULL CHECK (hit_count > 0),
    first_hit_at timestamptz NOT NULL,
    last_hit_at timestamptz NOT NULL,
    retired_at timestamptz,
    PRIMARY KEY (route_contract, consumer_id, build_id),
    CHECK (last_hit_at >= first_hit_at)
);

CREATE TABLE documents_reverse_reconciliations (
    reconciliation_id uuid PRIMARY KEY,
    source_generation bigint NOT NULL CHECK (source_generation > 0),
    target_generation bigint NOT NULL CHECK (target_generation > 0),
    source_count bigint NOT NULL CHECK (source_count >= 0),
    target_count bigint NOT NULL CHECK (target_count >= 0),
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256) = 32),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256) = 32),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('PENDING','ACCEPTED','REJECTED')),
    requested_by uuid NOT NULL,
    approved_by uuid,
    created_at timestamptz NOT NULL,
    decided_at timestamptz,
    CHECK (requested_by IS DISTINCT FROM approved_by),
    CHECK ((lifecycle_state = 'PENDING' AND approved_by IS NULL AND decided_at IS NULL) OR
           (lifecycle_state <> 'PENDING' AND approved_by IS NOT NULL AND decided_at IS NOT NULL))
);

