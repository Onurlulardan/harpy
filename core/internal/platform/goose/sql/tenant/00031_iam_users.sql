-- +goose Up
CREATE TABLE iam_users (
    id uuid NOT NULL PRIMARY KEY,
    revision bigint NOT NULL
);

-- +goose Down
DROP TABLE iam_users;
