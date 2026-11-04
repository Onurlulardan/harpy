-- +goose Up
CREATE TABLE work_tasks (
    id uuid PRIMARY KEY,
    board_id uuid NOT NULL,
    column_id uuid NOT NULL,
    creator_user_id uuid NOT NULL,
    task_kind text NOT NULL,
    status_category text NOT NULL,
    priority text NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    position bigint NOT NULL,
    row_version bigint NOT NULL,
    starts_at timestamptz,
    due_at timestamptz,
    completed_at timestamptz,
    cancelled_at timestamptz,
    deleted_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT work_tasks_kind_check CHECK (
        task_kind IN ('CUSTOM', 'FORM_FILL', 'FORM_REVIEW')
    ),
    CONSTRAINT work_tasks_status_check CHECK (
        status_category IN ('OPEN', 'IN_PROGRESS', 'IN_REVIEW', 'DONE', 'CANCELLED')
    ),
    CONSTRAINT work_tasks_priority_check CHECK (
        priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT')
    ),
    CONSTRAINT work_tasks_title_check CHECK (char_length(title) BETWEEN 1 AND 300),
    CONSTRAINT work_tasks_description_check CHECK (char_length(description) <= 16000),
    CONSTRAINT work_tasks_position_check CHECK (position > 0),
    CONSTRAINT work_tasks_row_version_check CHECK (row_version > 0)
);

-- +goose Down
DROP TABLE work_tasks;
