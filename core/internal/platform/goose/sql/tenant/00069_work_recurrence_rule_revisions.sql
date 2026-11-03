-- +goose Up
CREATE TABLE work_recurrence_rule_revisions (
    rule_id uuid NOT NULL PRIMARY KEY,
    revision bigint NOT NULL,
    contract_version bigint NOT NULL,
    dtstart text NOT NULL,
    timezone_name text NOT NULL,
    tzdb_version bigint NOT NULL,
    rrule text NOT NULL,
    exdates text NOT NULL,
    effective_from text NOT NULL,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE work_recurrence_rule_revisions;
