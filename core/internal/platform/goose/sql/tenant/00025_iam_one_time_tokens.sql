-- +goose Up
CREATE TABLE iam_one_time_tokens (
    id uuid NOT NULL PRIMARY KEY,
    state text NOT NULL,
    expires_at timestamptz
);

-- +goose Down
DROP TABLE iam_one_time_tokens;
