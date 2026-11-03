-- +goose Up
CREATE TABLE documents_restore_transient_dispositions (
    target_kind text NOT NULL PRIMARY KEY,
    target_id uuid NOT NULL
);

-- +goose Down
DROP TABLE documents_restore_transient_dispositions;
