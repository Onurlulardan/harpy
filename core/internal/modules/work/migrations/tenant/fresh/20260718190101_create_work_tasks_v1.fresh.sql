-- +goose Up
CREATE TABLE work_tasks (
    id uuid PRIMARY KEY,
    board_id uuid NOT NULL,
    column_id uuid NOT NULL,
    task_kind text NOT NULL CHECK (task_kind IN ('CUSTOM','FORM_FILL','FORM_REVIEW')),
    status_category text NOT NULL CHECK (status_category IN (
        'OPEN','IN_PROGRESS','IN_REVIEW','DONE','CANCELLED'
    )),
    title text NOT NULL CHECK (pg_catalog.octet_length(title) BETWEEN 1 AND 1200),
    description text NOT NULL CHECK (pg_catalog.octet_length(description) <= 64000),
    position bigint NOT NULL CHECK (position > 0),
    row_version bigint NOT NULL CHECK (row_version > 0),
    due_at timestamptz,
    completed_at timestamptz,
    cancelled_at timestamptz,
    deleted_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT work_tasks_board_fk FOREIGN KEY (board_id)
        REFERENCES work_boards(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT work_tasks_column_fk FOREIGN KEY (column_id)
        REFERENCES work_board_columns(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT work_tasks_position_key UNIQUE (column_id, position)
        DEFERRABLE INITIALLY DEFERRED,
    CHECK ((status_category = 'DONE') = (completed_at IS NOT NULL)),
    CHECK ((status_category = 'CANCELLED') = (cancelled_at IS NOT NULL))
);

CREATE TABLE work_task_assignees (
    task_id uuid NOT NULL,
    principal_kind text NOT NULL CHECK (principal_kind IN ('USER','ROLE','ORGANIZATION')),
    principal_id uuid NOT NULL,
    active boolean NOT NULL,
    removed_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (task_id, principal_kind, principal_id),
    CONSTRAINT work_task_assignees_task_fk FOREIGN KEY (task_id)
        REFERENCES work_tasks(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (active = (removed_at IS NULL))
);

CREATE TABLE work_labels (
    id uuid PRIMARY KEY,
    board_id uuid NOT NULL,
    name text NOT NULL CHECK (pg_catalog.octet_length(name) BETWEEN 1 AND 400),
    color text NOT NULL CHECK (color ~ '^#[0-9a-f]{6}$'),
    row_version bigint NOT NULL CHECK (row_version > 0),
    deleted_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT work_labels_board_fk FOREIGN KEY (board_id)
        REFERENCES work_boards(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE work_task_labels (
    task_id uuid NOT NULL,
    label_id uuid NOT NULL,
    active boolean NOT NULL,
    removed_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (task_id, label_id),
    CONSTRAINT work_task_labels_task_fk FOREIGN KEY (task_id)
        REFERENCES work_tasks(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT work_task_labels_label_fk FOREIGN KEY (label_id)
        REFERENCES work_labels(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (active = (removed_at IS NULL))
);

CREATE TABLE work_checklist_items (
    id uuid PRIMARY KEY,
    task_id uuid NOT NULL,
    item_text text NOT NULL CHECK (pg_catalog.octet_length(item_text) BETWEEN 1 AND 4000),
    completed boolean NOT NULL,
    position bigint NOT NULL CHECK (position > 0),
    row_version bigint NOT NULL CHECK (row_version > 0),
    deleted_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT work_checklist_items_task_fk FOREIGN KEY (task_id)
        REFERENCES work_tasks(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE UNIQUE INDEX work_labels_active_name_key
    ON work_labels(board_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX work_checklist_items_active_position_key
    ON work_checklist_items(task_id, position) WHERE deleted_at IS NULL;
CREATE INDEX work_tasks_board_list_idx ON work_tasks(board_id, column_id, position, id)
    WHERE deleted_at IS NULL;
CREATE INDEX work_tasks_due_idx ON work_tasks(due_at, id)
    WHERE due_at IS NOT NULL AND deleted_at IS NULL AND status_category NOT IN ('DONE','CANCELLED');
