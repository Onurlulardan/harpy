-- +goose Up
ALTER TABLE documents_uploads ADD COLUMN cancel_requested_at timestamptz;

GRANT UPDATE (cancel_requested_at) ON documents_uploads TO harpy_runtime;
GRANT INSERT (cancel_requested_at) ON documents_uploads TO harpy_adopter;
GRANT UPDATE (cancel_requested_at) ON documents_uploads TO harpy_adopter;
