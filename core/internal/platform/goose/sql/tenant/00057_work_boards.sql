-- +goose Up
CREATE TABLE work_boards (
    id uuid NOT NULL PRIMARY KEY,
    created_at timestamptz NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    creator_user_id uuid NOT NULL,
    lifecycle_state text NOT NULL,
    row_version bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    access_rank text NOT NULL,
    archived_at timestamptz,
    reference_epoch text NOT NULL
);

-- +goose Down
DROP TABLE work_boards;
