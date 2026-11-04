-- +goose Up
CREATE TABLE work_calendar_items (
    id uuid NOT NULL PRIMARY KEY,
    source_owner text NOT NULL,
    source_kind text NOT NULL,
    source_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    source_revision bigint NOT NULL,
    permission_revision bigint NOT NULL,
    event_sequence text NOT NULL,
    summary text NOT NULL,
    starts_at timestamptz,
    ends_at timestamptz,
    start_date text NOT NULL,
    end_date text NOT NULL,
    state text NOT NULL,
    visible_to_principals text NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE work_calendar_items;
