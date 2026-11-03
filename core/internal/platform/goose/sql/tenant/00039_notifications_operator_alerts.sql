-- +goose Up
CREATE TABLE notifications_operator_alerts (
    id uuid NOT NULL PRIMARY KEY,
    code text NOT NULL,
    resource_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    acknowledged_at timestamptz
);

-- +goose Down
DROP TABLE notifications_operator_alerts;
