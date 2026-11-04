-- +goose Up
CREATE TABLE notifications_template_versions (
    id uuid NOT NULL PRIMARY KEY,
    template_id uuid NOT NULL,
    locale text NOT NULL,
    template_kind text NOT NULL,
    subject_pattern text NOT NULL,
    body_pattern text NOT NULL,
    variables text NOT NULL,
    mandatory_email text NOT NULL,
    version_number bigint NOT NULL,
    content_digest text NOT NULL,
    published_at timestamptz
);

-- +goose Down
DROP TABLE notifications_template_versions;
