-- +goose Up
CREATE TABLE work_operations (
    operation_id uuid PRIMARY KEY,
    operation_kind text NOT NULL,
    actor_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    aggregate_kind text NOT NULL,
    state text NOT NULL CHECK (state IN ('CLAIMED','COMPLETED')),
    aggregate_id uuid,
    aggregate_version bigint CHECK (aggregate_version > 0),
    response_status integer CHECK (response_status BETWEEN 200 AND 299),
    response_document bytea CHECK (pg_catalog.octet_length(response_document) <= 65536),
    occurred_at timestamptz NOT NULL,
    completed_at timestamptz,
    CHECK ((state = 'CLAIMED' AND aggregate_id IS NULL AND aggregate_version IS NULL AND
            response_status IS NULL AND response_document IS NULL AND completed_at IS NULL) OR
           (state = 'COMPLETED' AND aggregate_id IS NOT NULL AND aggregate_version IS NOT NULL AND
            response_status IS NOT NULL AND response_document IS NOT NULL AND completed_at IS NOT NULL)),
    UNIQUE (actor_id, operation_kind, idempotency_key)
);

CREATE TABLE work_aggregate_sequences (
    aggregate_kind text NOT NULL,
    aggregate_id uuid NOT NULL,
    event_sequence bigint NOT NULL CHECK (event_sequence > 0),
    aggregate_version bigint NOT NULL CHECK (aggregate_version > 0),
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (aggregate_kind, aggregate_id)
);

CREATE TABLE work_audit_entries (
    audit_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    action text NOT NULL,
    aggregate_kind text NOT NULL,
    aggregate_id uuid NOT NULL,
    aggregate_version bigint NOT NULL CHECK (aggregate_version > 0),
    event_sequence bigint NOT NULL CHECK (event_sequence > 0),
    reason text,
    occurred_at timestamptz NOT NULL,
    CONSTRAINT work_audit_entries_operation_fk FOREIGN KEY (operation_id)
        REFERENCES work_operations(operation_id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    UNIQUE (aggregate_kind, aggregate_id, event_sequence)
);

CREATE TABLE work_purge_operations (
    operation_id uuid PRIMARY KEY,
    target_kind text NOT NULL,
    target_id uuid NOT NULL,
    expected_reference_epoch bigint NOT NULL CHECK (expected_reference_epoch > 0),
    state text NOT NULL CHECK (state IN (
        'COOLING','CANCELLED','EXECUTING','SUCCEEDED','FAILED'
    )),
    requested_by uuid NOT NULL,
    reason text NOT NULL CHECK (pg_catalog.octet_length(reason) BETWEEN 1 AND 4000),
    confirmation_name text NOT NULL,
    legal_hold boolean NOT NULL,
    blocker_count bigint NOT NULL CHECK (blocker_count >= 0),
    blocker_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(blocker_sha256) = 32),
    cooling_until timestamptz NOT NULL,
    destructive_started_at timestamptz,
    completed_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (target_kind, target_id, expected_reference_epoch)
);

CREATE TABLE work_purge_steps (
    operation_id uuid NOT NULL,
    step_no smallint NOT NULL CHECK (step_no > 0),
    step_kind text NOT NULL,
    expected_count bigint NOT NULL CHECK (expected_count >= 0),
    expected_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(expected_sha256) = 32),
    state text NOT NULL CHECK (state IN ('PENDING','RUNNING','SUCCEEDED','FAILED')),
    completed_at timestamptz,
    PRIMARY KEY (operation_id, step_no),
    CONSTRAINT work_purge_steps_operation_fk FOREIGN KEY (operation_id)
        REFERENCES work_purge_operations(operation_id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE work_id_tombstones (
    target_kind text NOT NULL,
    target_id uuid NOT NULL,
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 0),
    receipt_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(receipt_sha256) = 32),
    purged_at timestamptz NOT NULL,
    retain_until timestamptz NOT NULL,
    PRIMARY KEY (target_kind, target_id),
    CHECK (retain_until > purged_at)
);
