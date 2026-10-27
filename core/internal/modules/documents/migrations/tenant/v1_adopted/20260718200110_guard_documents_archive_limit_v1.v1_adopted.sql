-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_archive_entry_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE
    archive_state text;
    current_count bigint;
    current_bytes bigint;
BEGIN
    SELECT lifecycle_state INTO archive_state
    FROM documents_archives WHERE id = NEW.archive_id FOR UPDATE;
    IF archive_state IS DISTINCT FROM 'PENDING' THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'DOCUMENT_ARCHIVE_SNAPSHOT_CLOSED';
    END IF;
    SELECT count(*), COALESCE(sum(plaintext_bytes), 0)
      INTO current_count, current_bytes
      FROM documents_archive_entries WHERE archive_id = NEW.archive_id;
    IF current_count >= 10000 OR NEW.entry_order <> current_count OR
       current_bytes + NEW.plaintext_bytes > 10737418240 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'DOCUMENT_ARCHIVE_LIMIT_EXCEEDED';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER documents_archive_entries_limit_guard
BEFORE INSERT ON documents_archive_entries
FOR EACH ROW EXECUTE FUNCTION documents_guard_archive_entry_v1();
ALTER TABLE documents_archive_entries
    ENABLE ALWAYS TRIGGER documents_archive_entries_limit_guard;
