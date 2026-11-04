-- +goose Up
CREATE TABLE iam_permissions (
    id uuid NOT NULL PRIMARY KEY,
    resource_id uuid NOT NULL,
    action_id uuid NOT NULL,
    match_kind text NOT NULL,
    is_protected boolean NOT NULL,
    state text NOT NULL
);

-- +goose Down
DROP TABLE iam_permissions;
