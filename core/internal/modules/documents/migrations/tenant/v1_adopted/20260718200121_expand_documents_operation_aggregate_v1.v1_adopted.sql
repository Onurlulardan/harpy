-- +goose Up
ALTER TABLE documents_operations
    DROP CONSTRAINT documents_operations_aggregate_kind_check,
    ADD CONSTRAINT documents_operations_aggregate_kind_check
        CHECK (aggregate_kind IN (
            'FOLDER','ITEM','UPLOAD','ARCHIVE','LINK','LEGAL_HOLD'
        ));
