-- +goose Up
CREATE TABLE work_board_columns (
    id uuid NOT NULL PRIMARY KEY,
    board_id uuid NOT NULL,
    name text NOT NULL,
    color text NOT NULL,
    status_category text NOT NULL,
    position bigint NOT NULL,
    row_version bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE work_board_columns;
