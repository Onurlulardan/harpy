-- +goose Up
CREATE TABLE platform.billing_price_versions (
    id uuid NOT NULL PRIMARY KEY,
    offering_version_id uuid NOT NULL,
    currency text NOT NULL,
    cadence text NOT NULL,
    effective_start text NOT NULL,
    effective_end text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_price_versions;
