-- +goose Up
CREATE TABLE work_comments (
    task_id uuid NOT NULL,
    state text NOT NULL,
    id uuid NOT NULL PRIMARY KEY,
    created_at timestamptz NOT NULL,
    board_id uuid NOT NULL,
    author_id uuid NOT NULL,
    body jsonb NOT NULL,
    redaction_kind text NOT NULL,
    redaction_reason text NOT NULL,
    row_version bigint NOT NULL,
    redacted_at timestamptz,
    redacted_by text NOT NULL
);

-- +goose Down
DROP TABLE work_comments;
