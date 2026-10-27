-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_guard_publication_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_state text;
    v_resource text;
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'DELETE'
       OR TG_OP IS NOT DISTINCT FROM 'TRUNCATE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            CONSTRAINT = 'system_resource_layout_publications_lifecycle_guard',
            MESSAGE = 'system resource layout publication cannot be removed';
    END IF;
    SELECT revision.state, revision.resource_kind
    INTO v_state, v_resource
    FROM system_resource_layout_revisions AS revision
    WHERE revision.id = NEW.revision_id
    FOR KEY SHARE;
    IF NOT FOUND OR v_state IS DISTINCT FROM 'published'
       OR v_resource IS DISTINCT FROM NEW.resource_kind THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publications_target_guard',
            MESSAGE = 'system resource layout publication target is invalid';
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.generation IS DISTINCT FROM 1
           OR NEW.updated_at IS DISTINCT FROM NEW.created_at THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                CONSTRAINT = 'system_resource_layout_publications_lifecycle_guard',
                MESSAGE = 'system resource layout publication must begin at one';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS DISTINCT FROM 'UPDATE'
       OR NEW.resource_kind IS DISTINCT FROM OLD.resource_kind
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.generation IS DISTINCT FROM OLD.generation + 1
       OR NEW.updated_at < OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publications_lifecycle_guard',
            MESSAGE = 'system resource layout publication update is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_guard_publication_v1()
    FROM PUBLIC;

CREATE TRIGGER system_resource_layout_publications_write_guard
BEFORE INSERT OR UPDATE ON system_resource_layout_publications
FOR EACH ROW EXECUTE FUNCTION system_resource_layout_guard_publication_v1();
ALTER TABLE system_resource_layout_publications
    ENABLE ALWAYS TRIGGER system_resource_layout_publications_write_guard;

CREATE TRIGGER system_resource_layout_publications_delete_guard
BEFORE DELETE ON system_resource_layout_publications
FOR EACH ROW EXECUTE FUNCTION system_resource_layout_guard_publication_v1();
ALTER TABLE system_resource_layout_publications
    ENABLE ALWAYS TRIGGER system_resource_layout_publications_delete_guard;

CREATE TRIGGER system_resource_layout_publications_truncate_guard
BEFORE TRUNCATE ON system_resource_layout_publications
FOR EACH STATEMENT EXECUTE FUNCTION system_resource_layout_guard_publication_v1();
ALTER TABLE system_resource_layout_publications
    ENABLE ALWAYS TRIGGER system_resource_layout_publications_truncate_guard;

