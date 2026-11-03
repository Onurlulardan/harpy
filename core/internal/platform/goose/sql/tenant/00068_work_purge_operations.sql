-- +goose Up
CREATE TABLE work_purge_operations (
    operation_id uuid NOT NULL PRIMARY KEY
);

-- +goose Down
DROP TABLE work_purge_operations;
