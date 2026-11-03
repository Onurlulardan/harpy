-- +goose Up
CREATE TABLE iam_resources (
    id uuid NOT NULL PRIMARY KEY,
    resource_key text NOT NULL,
    state text NOT NULL
);

-- +goose Down
DROP TABLE iam_resources;
