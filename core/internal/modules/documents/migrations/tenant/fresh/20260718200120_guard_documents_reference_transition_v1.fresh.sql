-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_reference_transition_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    IF TG_OP = 'DELETE' OR OLD.lifecycle_state <> 'ACTIVE' OR
       NEW.lifecycle_state <> 'RELEASED' OR OLD.released_at IS NOT NULL OR
       NEW.released_at IS NULL OR
       (pg_catalog.to_jsonb(NEW) - 'lifecycle_state' - 'released_at')
       IS DISTINCT FROM
       (pg_catalog.to_jsonb(OLD) - 'lifecycle_state' - 'released_at') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENT_REFERENCE_IMMUTABLE';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER documents_links_transition_guard
BEFORE UPDATE OR DELETE ON documents_links
FOR EACH ROW EXECUTE FUNCTION documents_guard_reference_transition_v1();
ALTER TABLE documents_links ENABLE ALWAYS TRIGGER documents_links_transition_guard;

CREATE TRIGGER documents_legal_holds_transition_guard
BEFORE UPDATE OR DELETE ON documents_legal_holds
FOR EACH ROW EXECUTE FUNCTION documents_guard_reference_transition_v1();
ALTER TABLE documents_legal_holds
    ENABLE ALWAYS TRIGGER documents_legal_holds_transition_guard;

GRANT EXECUTE ON FUNCTION documents_guard_reference_transition_v1()
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION documents_guard_reference_transition_v1()
TO forgecrud_adopter;
