-- +goose Up
CREATE TABLE documents_dr_holds (
    lifecycle_state text NOT NULL PRIMARY KEY
);

-- +goose Down
DROP TABLE documents_dr_holds;
