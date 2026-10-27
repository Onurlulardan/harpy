-- +goose Up
ALTER TABLE work_tasks
    ADD COLUMN IF NOT EXISTS creator_user_id uuid,
    ADD COLUMN IF NOT EXISTS priority text NOT NULL DEFAULT 'MEDIUM',
    ADD COLUMN IF NOT EXISTS starts_at timestamptz;

ALTER TABLE work_board_columns
    ADD COLUMN IF NOT EXISTS color text NOT NULL DEFAULT '#64748b';

ALTER TABLE work_calendar_events
    ADD COLUMN IF NOT EXISTS color text NOT NULL DEFAULT '#64748b';

UPDATE work_tasks AS task
SET creator_user_id = board.creator_user_id
FROM work_boards AS board
WHERE task.board_id = board.id
  AND task.creator_user_id IS NULL;

ALTER TABLE work_tasks
    ALTER COLUMN creator_user_id SET NOT NULL,
    ADD CONSTRAINT work_tasks_priority_check
        CHECK (priority IN ('LOW','MEDIUM','HIGH','URGENT'));

ALTER TABLE work_board_columns
    ADD CONSTRAINT work_board_columns_color_check
        CHECK (color ~ '^#[0-9a-f]{6}$');

ALTER TABLE work_calendar_events
    ADD CONSTRAINT work_calendar_events_color_check
        CHECK (color ~ '^#[0-9a-f]{6}$');

CREATE INDEX work_tasks_priority_due_idx
    ON work_tasks(priority, due_at, id)
    WHERE deleted_at IS NULL;
