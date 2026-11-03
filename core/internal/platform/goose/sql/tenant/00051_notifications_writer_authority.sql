-- +goose Up
CREATE TABLE notifications_writer_authority (
    writer text NOT NULL PRIMARY KEY,
    lifecycle_state text NOT NULL,
    generation text NOT NULL,
    row_version bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    singleton boolean NOT NULL
);

-- +goose Down
DROP TABLE notifications_writer_authority;
