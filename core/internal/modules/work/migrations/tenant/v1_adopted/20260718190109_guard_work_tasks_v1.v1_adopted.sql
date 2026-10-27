-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_guard_task_board_limit_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE column_board uuid; column_total bigint; board_total bigint;
BEGIN
    PERFORM 1 FROM work_boards WHERE id = NEW.board_id FOR UPDATE;
    SELECT board_id INTO column_board FROM work_board_columns
        WHERE id = NEW.column_id FOR UPDATE;
    IF column_board IS DISTINCT FROM NEW.board_id THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'WORK_CROSS_BOARD_REFERENCE';
    END IF;
    IF TG_OP = 'INSERT' THEN
        SELECT count(*) INTO column_total FROM work_tasks
            WHERE column_id = NEW.column_id AND deleted_at IS NULL;
        SELECT count(*) INTO board_total FROM work_tasks
            WHERE board_id = NEW.board_id AND deleted_at IS NULL;
    ELSE
        IF OLD.column_id IS DISTINCT FROM NEW.column_id THEN
            SELECT count(*) INTO column_total FROM work_tasks
                WHERE column_id = NEW.column_id AND deleted_at IS NULL;
        END IF;
        IF OLD.board_id IS DISTINCT FROM NEW.board_id THEN
            SELECT count(*) INTO board_total FROM work_tasks
                WHERE board_id = NEW.board_id AND deleted_at IS NULL;
        END IF;
    END IF;
    IF column_total >= 10000 OR board_total >= 50000 THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'WORK_BOARD_LIMIT_EXCEEDED';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER work_tasks_limit_guard
BEFORE INSERT OR UPDATE OF board_id, column_id ON work_tasks
FOR EACH ROW EXECUTE FUNCTION work_guard_task_board_limit_v1();
ALTER TABLE work_tasks ENABLE ALWAYS TRIGGER work_tasks_limit_guard;
