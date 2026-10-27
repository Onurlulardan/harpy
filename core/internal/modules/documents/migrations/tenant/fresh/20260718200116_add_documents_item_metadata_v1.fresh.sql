-- +goose Up
ALTER TABLE documents_items
    ADD COLUMN description text NOT NULL DEFAULT ''
        CHECK (pg_catalog.octet_length(description) <= 4096),
    ADD COLUMN tags jsonb NOT NULL DEFAULT '[]'::jsonb
        CHECK (pg_catalog.jsonb_typeof(tags) = 'array' AND
               pg_catalog.octet_length(tags::text) <= 2048);

ALTER TABLE documents_uploads
    ADD COLUMN description text NOT NULL DEFAULT ''
        CHECK (pg_catalog.octet_length(description) <= 4096),
    ADD COLUMN tags jsonb NOT NULL DEFAULT '[]'::jsonb
        CHECK (pg_catalog.jsonb_typeof(tags) = 'array' AND
               pg_catalog.octet_length(tags::text) <= 2048);

GRANT INSERT (description, tags) ON documents_items TO harpy_runtime;
GRANT UPDATE (description, tags) ON documents_items TO harpy_runtime;
GRANT INSERT (description, tags) ON documents_uploads TO harpy_runtime;
GRANT INSERT (description, tags) ON documents_items TO harpy_adopter;
GRANT UPDATE (description, tags) ON documents_items TO harpy_adopter;
