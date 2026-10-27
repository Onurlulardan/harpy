-- +goose Up
CREATE TABLE work_v1_adoption_stage (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    source_kind text NOT NULL CHECK (source_kind IN (
        'task_boards','task_board_columns','board_members','board_tasks',
        'board_task_comments','calendar_events'
    )),
    source_id uuid NOT NULL,
    target_kind text NOT NULL CHECK (target_kind IN (
        'BOARD','COLUMN','MEMBERSHIP','TASK','ASSIGNEE','LABEL','TASK_LABEL',
        'CHECKLIST','COMMENT','CALENDAR','RECURRENCE','FORM_LINK','WORKFLOW_PROJECTION'
    )),
    target_id uuid NOT NULL,
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256) = 32),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256) = 32),
    target_record jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(target_record) = 'object'),
    removed boolean NOT NULL,
    source_xid xid8 NOT NULL,
    commit_lsn pg_lsn NOT NULL,
    batch_ordinal integer NOT NULL CHECK (batch_ordinal BETWEEN 1 AND 2147483),
    change_ordinal integer NOT NULL CHECK (change_ordinal BETWEEN 1 AND 1000),
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id, run_id, target_kind, target_id)
);

CREATE INDEX work_v1_adoption_stage_source_idx
    ON work_v1_adoption_stage (
        operation_id, run_id, source_kind, source_id, target_kind, target_id
    );

REVOKE UPDATE (
    state, run_id, generation, reverse_reconciliation_id, updated_by, updated_at
) ON work_v1_writer_fence FROM harpy_runtime;
REVOKE INSERT (
    run_id, cutover_generation, manifest_sha256, transform_version, state,
    source_count, target_count, source_sha256, target_sha256, final_lsn,
    applied_lsn, acceptance_sha256, created_at, updated_at
) ON work_adoption_runs FROM harpy_runtime;
REVOKE UPDATE (
    state, source_count, target_count, source_sha256, target_sha256,
    final_lsn, applied_lsn, acceptance_sha256, updated_at
) ON work_adoption_runs FROM harpy_runtime;
REVOKE INSERT (
    run_id, source_kind, source_id, target_kind, target_id, source_sha256,
    target_sha256, applied_lsn, state, issue_code, updated_at
) ON work_adoption_objects FROM harpy_runtime;
REVOKE UPDATE (
    source_sha256, target_sha256, applied_lsn, state, issue_code, updated_at
) ON work_adoption_objects FROM harpy_runtime;
REVOKE INSERT (
    run_id, issue_id, source_kind, source_id, issue_code, evidence_sha256, created_at
) ON work_adoption_issues FROM harpy_runtime;
REVOKE INSERT (
    reconciliation_id, source_generation, target_generation, source_count,
    target_count, source_sha256, target_sha256, state, requested_by, approved_by,
    created_at, decided_at
) ON work_reverse_reconciliations FROM harpy_runtime;
REVOKE UPDATE (state, approved_by, decided_at)
ON work_reverse_reconciliations FROM harpy_runtime;
