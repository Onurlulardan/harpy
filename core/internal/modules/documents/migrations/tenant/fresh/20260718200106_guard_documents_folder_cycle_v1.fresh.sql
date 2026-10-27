-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_folder_cycle_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE ancestor_id uuid;
BEGIN
    LOCK TABLE documents_folders IN SHARE ROW EXCLUSIVE MODE;
    IF NEW.parent_id IS NULL THEN
        RETURN NEW;
    END IF;
    ancestor_id := NEW.parent_id;
    WHILE ancestor_id IS NOT NULL LOOP
        IF ancestor_id = NEW.id THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'DOCUMENT_FOLDER_CYCLE';
        END IF;
        SELECT parent_id INTO ancestor_id
        FROM documents_folders WHERE id = ancestor_id;
    END LOOP;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER documents_folders_cycle_guard
BEFORE INSERT OR UPDATE OF parent_id ON documents_folders
FOR EACH ROW EXECUTE FUNCTION documents_guard_folder_cycle_v1();
ALTER TABLE documents_folders ENABLE ALWAYS TRIGGER documents_folders_cycle_guard;
