-- +goose Up
CREATE TABLE platform.billing_offerings (
    id uuid NOT NULL PRIMARY KEY,
    lifecycle text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_offerings;
