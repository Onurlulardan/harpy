-- +goose Up
CREATE TABLE work_adoption_runs (
    run_id uuid PRIMARY KEY,
    cutover_generation bigint NOT NULL CHECK (cutover_generation > 0),
    manifest_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(manifest_sha256) = 32),
    transform_version text NOT NULL CHECK (transform_version = 'work_adoption_v1'),
    state text NOT NULL CHECK (state IN (
        'INVENTORY','COPYING','RECONCILING','ACCEPTED','BLOCKED','REVERSED'
    )),
    source_count bigint NOT NULL CHECK (source_count >= 0),
    target_count bigint NOT NULL CHECK (target_count >= 0),
    source_sha256 bytea CHECK (source_sha256 IS NULL OR pg_catalog.octet_length(source_sha256) = 32),
    target_sha256 bytea CHECK (target_sha256 IS NULL OR pg_catalog.octet_length(target_sha256) = 32),
    final_lsn pg_lsn,
    applied_lsn pg_lsn,
    acceptance_sha256 bytea CHECK (acceptance_sha256 IS NULL OR pg_catalog.octet_length(acceptance_sha256) = 32),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE work_adoption_objects (
    run_id uuid NOT NULL,
    source_kind text NOT NULL CHECK (source_kind IN (
        'task_boards','task_board_columns','board_members','board_tasks',
        'board_task_comments','calendar_events'
    )),
    source_id uuid NOT NULL,
    target_kind text NOT NULL,
    target_id uuid NOT NULL,
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256) = 32),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256) = 32),
    applied_lsn pg_lsn,
    state text NOT NULL CHECK (state IN ('COPIED','MATCHED','BLOCKED')),
    issue_code text,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (run_id, source_kind, source_id, target_kind, target_id),
    CONSTRAINT work_adoption_objects_run_fk FOREIGN KEY (run_id)
        REFERENCES work_adoption_runs(run_id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK ((state = 'BLOCKED') = (issue_code IS NOT NULL))
);

CREATE TABLE work_adoption_issues (
    run_id uuid NOT NULL,
    issue_id uuid NOT NULL,
    source_kind text NOT NULL,
    source_id uuid,
    issue_code text NOT NULL CHECK (issue_code IN (
        'INVALID_ENUM','CROSS_BOARD_REFERENCE','ORPHAN','UNRESOLVED_FORM',
        'INVALID_RECURRENCE','CHECKSUM_MISMATCH','FUTURE_WORK_PENDING'
    )),
    evidence_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(evidence_sha256) = 32),
    created_at timestamptz NOT NULL,
    PRIMARY KEY (run_id, issue_id),
    CONSTRAINT work_adoption_issues_run_fk FOREIGN KEY (run_id)
        REFERENCES work_adoption_runs(run_id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE work_compatibility_hits (
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

CREATE TABLE work_reverse_reconciliations (
    reconciliation_id uuid PRIMARY KEY,
    source_generation bigint NOT NULL CHECK (source_generation > 0),
    target_generation bigint NOT NULL CHECK (target_generation > 0),
    source_count bigint NOT NULL CHECK (source_count >= 0),
    target_count bigint NOT NULL CHECK (target_count >= 0),
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256) = 32),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256) = 32),
    state text NOT NULL CHECK (state IN ('PENDING','ACCEPTED','REJECTED')),
    requested_by uuid NOT NULL,
    approved_by uuid,
    created_at timestamptz NOT NULL,
    decided_at timestamptz,
    CHECK (requested_by IS DISTINCT FROM approved_by),
    CHECK ((state = 'PENDING' AND approved_by IS NULL AND decided_at IS NULL) OR
           (state <> 'PENDING' AND approved_by IS NOT NULL AND decided_at IS NOT NULL))
);
