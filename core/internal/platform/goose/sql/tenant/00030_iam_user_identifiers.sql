-- +goose Up
CREATE TABLE iam_user_identifiers (
    id uuid NOT NULL PRIMARY KEY,
    revision bigint NOT NULL,
    state text NOT NULL,
    retired_at timestamptz,
    cooldown_until text NOT NULL,
    normalized_value text NOT NULL,
    user_id uuid NOT NULL,
    kind text NOT NULL
);

-- +goose Down
DROP TABLE iam_user_identifiers;
