-- +goose Up
CREATE TABLE platform.billing_discount_offerings (
    discount_version_id uuid NOT NULL PRIMARY KEY,
    offering_version_id uuid NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_discount_offerings;
