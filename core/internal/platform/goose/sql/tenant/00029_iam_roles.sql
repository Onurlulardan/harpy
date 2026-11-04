-- +goose Up
CREATE TABLE iam_roles (
    id uuid NOT NULL PRIMARY KEY,
    revision bigint NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    name_key text NOT NULL,
    status text NOT NULL,
    is_protected boolean NOT NULL
);

-- +goose Down
DROP TABLE iam_roles;
