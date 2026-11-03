-- +goose Up
CREATE TABLE work_calendar_events (
    id uuid NOT NULL PRIMARY KEY,
    owner_id uuid NOT NULL,
    audience text NOT NULL,
    board_id uuid NOT NULL,
    organization_id uuid NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    starts_at timestamptz,
    ends_at timestamptz,
    start_date text NOT NULL,
    end_date text NOT NULL,
    timezone_name text NOT NULL,
    rule_id uuid NOT NULL,
    rule_revision bigint NOT NULL,
    occurrence_start text NOT NULL,
    tzdb_version bigint NOT NULL,
    generated text NOT NULL,
    generated_state text NOT NULL,
    edited text NOT NULL,
    live_reference_count bigint NOT NULL,
    legal_hold text NOT NULL,
    row_version bigint NOT NULL,
    deleted_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    color text NOT NULL
);

-- +goose Down
DROP TABLE work_calendar_events;
