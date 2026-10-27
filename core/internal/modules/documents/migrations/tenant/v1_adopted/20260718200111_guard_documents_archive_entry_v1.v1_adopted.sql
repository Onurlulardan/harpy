-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_archive_entry_immutable_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '42501',
        MESSAGE = 'DOCUMENT_ARCHIVE_ENTRY_IMMUTABLE';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER documents_archive_entries_immutable_guard
BEFORE UPDATE OR DELETE ON documents_archive_entries
FOR EACH ROW EXECUTE FUNCTION documents_guard_archive_entry_immutable_v1();
ALTER TABLE documents_archive_entries
    ENABLE ALWAYS TRIGGER documents_archive_entries_immutable_guard;

