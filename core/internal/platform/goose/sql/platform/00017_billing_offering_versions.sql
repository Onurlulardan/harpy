-- +goose Up
CREATE TABLE platform.billing_offering_versions (
    id uuid NOT NULL PRIMARY KEY,
    trial_days text NOT NULL,
    offering_id uuid NOT NULL,
    valid_start text NOT NULL,
    valid_end text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_offering_versions;
