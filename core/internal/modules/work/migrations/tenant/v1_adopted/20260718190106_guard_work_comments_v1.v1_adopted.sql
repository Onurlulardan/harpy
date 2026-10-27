-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_guard_comment_mutation_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'WORK_COMMENT_IMMUTABLE';
    END IF;
    IF OLD.state <> 'VISIBLE' OR NEW.state <> 'REDACTED'
       OR NEW.body <> '' OR NEW.redaction_kind IS NULL OR NEW.redacted_by IS NULL
       OR NEW.redacted_at IS NULL OR NEW.row_version <> OLD.row_version + 1
       OR (NEW.id, NEW.board_id, NEW.task_id, NEW.author_id, NEW.created_at)
          IS DISTINCT FROM
          (OLD.id, OLD.board_id, OLD.task_id, OLD.author_id, OLD.created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'WORK_COMMENT_IMMUTABLE';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER work_comments_mutation_guard
BEFORE UPDATE OR DELETE ON work_comments
FOR EACH ROW EXECUTE FUNCTION work_guard_comment_mutation_v1();
ALTER TABLE work_comments ENABLE ALWAYS TRIGGER work_comments_mutation_guard;
