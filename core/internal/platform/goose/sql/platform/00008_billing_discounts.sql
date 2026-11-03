-- +goose Up
CREATE TABLE platform.billing_discounts (
    normalized_code text NOT NULL,
    lifecycle text NOT NULL,
    authority_id uuid NOT NULL,
    id uuid NOT NULL PRIMARY KEY
);

-- +goose Down
DROP TABLE platform.billing_discounts;
