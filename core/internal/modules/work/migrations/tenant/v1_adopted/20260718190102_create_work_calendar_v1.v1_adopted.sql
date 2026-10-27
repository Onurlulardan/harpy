-- +goose Up
CREATE TABLE work_comments (
    id uuid PRIMARY KEY,
    board_id uuid NOT NULL,
    task_id uuid NOT NULL,
    author_id uuid NOT NULL,
    body text NOT NULL CHECK (pg_catalog.octet_length(body) <= 64000),
    state text NOT NULL CHECK (state IN ('VISIBLE','REDACTED')),
    redaction_kind text CHECK (redaction_kind IN ('AUTHOR','MODERATOR')),
    redacted_by uuid,
    redaction_reason text,
    row_version bigint NOT NULL CHECK (row_version > 0),
    created_at timestamptz NOT NULL,
    redacted_at timestamptz,
    CONSTRAINT work_comments_board_fk FOREIGN KEY (board_id)
        REFERENCES work_boards(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT work_comments_task_fk FOREIGN KEY (task_id)
        REFERENCES work_tasks(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK ((state = 'VISIBLE' AND body <> '' AND redaction_kind IS NULL AND
            redacted_by IS NULL AND redacted_at IS NULL) OR
           (state = 'REDACTED' AND body = '' AND redaction_kind IS NOT NULL AND
            redacted_by IS NOT NULL AND redacted_at IS NOT NULL)),
    CHECK (redaction_kind <> 'MODERATOR' OR pg_catalog.octet_length(redaction_reason) BETWEEN 1 AND 4000)
);

CREATE TABLE work_recurrence_rules (
    id uuid PRIMARY KEY,
    owner_event_id uuid NOT NULL UNIQUE,
    current_revision bigint NOT NULL CHECK (current_revision > 0),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('ACTIVE','RETIRED')),
    row_version bigint NOT NULL CHECK (row_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE work_recurrence_rule_revisions (
    rule_id uuid NOT NULL,
    revision bigint NOT NULL CHECK (revision > 0),
    contract_version text NOT NULL CHECK (contract_version = 'work_rrule_v1'),
    dtstart text NOT NULL,
    timezone_name text NOT NULL,
    tzdb_version text NOT NULL,
    rrule text NOT NULL,
    exdates text[] NOT NULL CHECK (cardinality(exdates) <= 500),
    effective_from timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (rule_id, revision),
    CONSTRAINT work_recurrence_rule_revisions_rule_fk FOREIGN KEY (rule_id)
        REFERENCES work_recurrence_rules(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE work_calendar_events (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL,
    audience text NOT NULL CHECK (audience IN ('PRIVATE','BOARD','ORGANIZATION')),
    board_id uuid,
    organization_id uuid,
    title text NOT NULL CHECK (pg_catalog.octet_length(title) BETWEEN 1 AND 1200),
    description text NOT NULL CHECK (pg_catalog.octet_length(description) <= 64000),
    starts_at timestamptz,
    ends_at timestamptz,
    start_date date,
    end_date date,
    timezone_name text NOT NULL,
    rule_id uuid,
    rule_revision bigint,
    occurrence_start timestamptz,
    tzdb_version text,
    generated boolean NOT NULL,
    generated_state text CHECK (generated_state IN ('ACTIVE','CANCELLED','CONTENT_PURGED')),
    edited boolean NOT NULL,
    live_reference_count bigint NOT NULL CHECK (live_reference_count >= 0),
    legal_hold boolean NOT NULL,
    row_version bigint NOT NULL CHECK (row_version > 0),
    deleted_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT work_calendar_events_board_fk FOREIGN KEY (board_id)
        REFERENCES work_boards(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT work_calendar_events_rule_revision_fk FOREIGN KEY (rule_id, rule_revision)
        REFERENCES work_recurrence_rule_revisions(rule_id, revision)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK ((starts_at IS NOT NULL AND ends_at > starts_at AND start_date IS NULL AND end_date IS NULL) OR
           (starts_at IS NULL AND ends_at IS NULL AND start_date IS NOT NULL AND end_date > start_date)),
    CHECK ((audience = 'PRIVATE' AND board_id IS NULL AND organization_id IS NULL) OR
           (audience = 'BOARD' AND board_id IS NOT NULL AND organization_id IS NULL) OR
           (audience = 'ORGANIZATION' AND board_id IS NULL AND organization_id IS NOT NULL)),
    CHECK ((NOT generated AND rule_id IS NULL AND rule_revision IS NULL AND occurrence_start IS NULL AND
            tzdb_version IS NULL AND generated_state IS NULL) OR
           (generated AND rule_id IS NOT NULL AND rule_revision IS NOT NULL AND occurrence_start IS NOT NULL AND
            tzdb_version IS NOT NULL AND generated_state IS NOT NULL))
);

ALTER TABLE work_recurrence_rules ADD CONSTRAINT work_recurrence_rules_event_fk
    FOREIGN KEY (owner_event_id) REFERENCES work_calendar_events(id)
    ON UPDATE NO ACTION ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;
CREATE UNIQUE INDEX work_calendar_occurrence_key
    ON work_calendar_events(rule_id, rule_revision, occurrence_start) WHERE generated;
CREATE INDEX work_calendar_range_timed_idx ON work_calendar_events(starts_at, ends_at, id)
    WHERE deleted_at IS NULL AND starts_at IS NOT NULL;
CREATE INDEX work_calendar_range_date_idx ON work_calendar_events(start_date, end_date, id)
    WHERE deleted_at IS NULL AND start_date IS NOT NULL;
