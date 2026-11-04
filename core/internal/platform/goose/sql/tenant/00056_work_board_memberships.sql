-- +goose Up
CREATE TABLE work_board_memberships (
    id uuid NOT NULL PRIMARY KEY,
    board_id uuid NOT NULL,
    principal_kind text NOT NULL,
    principal_id uuid NOT NULL,
    board_role text NOT NULL,
    row_version bigint NOT NULL,
    created_at timestamptz NOT NULL,
    active text NOT NULL,
    updated_at timestamptz NOT NULL,
    deactivated_at timestamptz
);

-- +goose Down
DROP TABLE work_board_memberships;
