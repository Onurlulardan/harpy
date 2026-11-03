-- +goose Up
CREATE TABLE work_labels (
    name text NOT NULL,
    id uuid NOT NULL PRIMARY KEY,
    deleted_at timestamptz,
    board_id uuid NOT NULL,
    color text NOT NULL,
    row_version bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE work_labels;
