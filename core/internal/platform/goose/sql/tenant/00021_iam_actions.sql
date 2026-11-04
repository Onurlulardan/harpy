-- +goose Up
CREATE TABLE iam_actions (
    id uuid NOT NULL PRIMARY KEY,
    action_key text NOT NULL,
    state text NOT NULL
);

-- +goose Down
DROP TABLE iam_actions;
