-- +goose Up
CREATE TABLE iam_organizations (
    id uuid NOT NULL PRIMARY KEY,
    status text NOT NULL,
    revision bigint NOT NULL,
    slug text NOT NULL
);

-- +goose Down
DROP TABLE iam_organizations;
